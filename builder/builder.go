// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sync"
	"time"

	"bldy.build/build/executor"
	"bldy.build/build/namespace"
	"github.com/pkg/errors"

	"sevki.org/pqueue"

	"strings"

	"bldy.build/build"
	"bldy.build/build/graph"
)

type Update struct {
	TimeStamp time.Time
	Target    string
	Status    build.Status
	Worker    int
	Cached    bool
}

type Builder struct {
	Origin      string
	Wd          string
	ProjectPath string
	Total       int
	Timeout     chan bool `json:"-"`
	ptr         *graph.Node
	graph       *graph.Graph
	pq          *pqueue.PQueue `json:"-"`
	config      *Config
	notifier    Notifier `json:"-"`
	start       time.Time

	wg sync.WaitGroup
}

type Notifier interface {
	Update(*graph.Node)
	Error(error)
	Done(time.Duration)
}

type Config struct {
	Fresh    bool
	BuildOut *string
	Cache    *string
}

func New(g *graph.Graph, c *Config, n Notifier) (b Builder) {
	var err error
	b.Wd, err = os.Getwd()
	if err != nil {
		l.Fatal(err)
	}
	b.pq = pqueue.New()
	b.graph = g
	b.notifier = n
	if c.Fresh {
		tmpDir, err := ioutil.TempDir("", fmt.Sprintf("bldy_tmp_%s_", g.Root.Label.Name()))
		if err != nil {
			log.Fatal(err)
		}
		c.Cache = &tmpDir
	} else if c.Cache == nil {
		c.Cache = bldyCache()
	}
	if c.BuildOut == nil {
		x := path.Join(g.Workspace().AbsPath(), "build_out")
		c.BuildOut = &x
	}

	b.config = c
	b.ProjectPath = g.Workspace().AbsPath()

	return
}

var (
	l = log.New(os.Stdout, "builder: ", 0)
)

func bldyCache() *string {
	usr, err := user.Current()
	if err != nil {
		l.Fatal(err)
	}
	x := path.Join(usr.HomeDir, "/.cache/bldy")
	return &x
}

func (b *Builder) Execute(ctx context.Context, r int) {
	b.start = time.Now()
	for i := 0; i < r; i++ {
		go b.work(ctx, i)
	}
	if b.graph == nil {
		l.Fatal("couldn't find the build graph")
	}
	if b.graph.Root == nil {
		l.Fatal("couldn't find the graph root")
	}
	b.wg.Add(1)
	b.visit(b.graph.Root)
	b.wg.Wait()
}

func (b *Builder) build(ctx context.Context, n *graph.Node) error {
	executor := executor.New(ctx, b.buildpath(n))
	n.Start = time.Now().UnixNano()

	err := n.Target.Build(executor)

	n.End = time.Now().UnixNano()

	n.Status = build.Fail
	if err == nil {
		n.Status = build.Success
	}

	n.Output = executor.CombinedLog()

	b.saveLog(n)
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) work(ctx context.Context, workerNumber int) {
	for {
		job := b.pq.Pop().(*graph.Node)

		job.Worker = fmt.Sprintf("%d", workerNumber)

		if job.Status != build.Pending {
			continue
		}
		job.Lock()

		job.Status = build.Building

		b.notifier.Update(job)

		if b.cached(job) {
			if err := b.builderror(job); err != nil {
				b.notifier.Update(job)
				b.notifier.Error(err)
			}
		} else {
			if err := b.prepare(job); err != nil {
				b.notifier.Update(job)
				b.notifier.Error(errors.Wrap(err, "build"))
			}

			if err := ctx.Err(); err != nil {
				b.notifier.Update(job)
				b.notifier.Error(err)
			}

			if err := b.build(ctx, job); err != nil {
				b.notifier.Update(job)
				b.notifier.Error(err)

			}
		}
		job.Once.Do(func() {
			for _, parent := range job.Parents {
				parent.WG.Done()
			}
		})

		b.notifier.Update(job)

		if job.IsRoot {
			if job.Status == build.Success {
				b.install(job)
			}
			b.notifier.Done(time.Now().Sub(b.start))
			b.wg.Done()
		}
		job.Unlock()
	}

}

func (b *Builder) visit(n *graph.Node) {
	// This is not an airplane so let's make sure children get their masks on before the parents.
	for _, child := range n.Children {
		// Visit children first
		go b.visit(child)
	}
	n.WG.Wait()
	n.Priority()
	b.pq.Push(n)
}

func (b *Builder) install(job *graph.Node) error {
	if err := os.MkdirAll(
		*b.config.BuildOut,
		os.ModeDir|os.ModePerm,
	); err != nil {
		l.Fatalf("copying job %s failed: %s", job.Target.Name(), err.Error())
	}

	for _, output := range job.Target.Outputs() {
		target := filepath.Base(output)
		targetDir := strings.TrimRight(output, target)
		buildOutTarget := filepath.Join(
			*b.config.BuildOut,
			targetDir,
		)

		src := filepath.Join(
			b.buildpath(job),
			output,
		)
		dst := filepath.Join(
			buildOutTarget,
			target,
		)
		if err := os.MkdirAll(
			buildOutTarget,
			os.ModeDir|os.ModePerm,
		); err != nil {
			l.Fatalf("linking job %s failed: %s", job.Target.Name(), err.Error())
		}

		dstp := filepath.Join(
			buildOutTarget,
			target,
		)
		d, err := filepath.EvalSymlinks(src)
		stat, err := os.Stat(d)

		if os.IsNotExist(err) {
			return fmt.Errorf("cannot install %s: file %s doesn't exist", job.Target.Name(), src)
		}
		in, err := os.Open(d)
		if err != nil {
			l.Fatalf("copy: can't finiliaze %s. copying %q to %q failed: %s\n", job.Target.Name(), src, dstp, err)
		}

		out, err := os.OpenFile(dstp, os.O_RDWR|os.O_CREATE, stat.Mode())
		if err != nil {
			l.Fatal(err)
		}

		if _, err := io.Copy(out, in); err != nil {
			l.Fatalf("copy: can't finiliaze %s. copying from %q to %q failed: %s\n", job.Target.Name(), src, dst, err)
		}
		if err := in.Close(); err != nil {
			l.Fatal(err)
		}
		if err := out.Close(); err != nil {
			l.Fatal(err)
		}
	}

	return nil
}

func (b *Builder) createOutputDirs(n *graph.Node) error {
	for _, output := range n.Target.Outputs() {
		target := filepath.Base(output)
		targetDir := strings.TrimRight(output, target)
		outputDir := filepath.Join(
			b.buildpath(n),
			targetDir,
		)

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) bindChildOutputs(n *graph.Node) error {
	for _, c := range n.Children {
		for _, output := range c.Target.Outputs() {

			src := filepath.Join(
				b.buildpath(c),
				output,
			)
			dst := filepath.Join(
				b.buildpath(n),
				output,
			)
			target := filepath.Base(output)
			targetDir := strings.TrimRight(output, target)
			outputDir := filepath.Join(
				b.buildpath(n),
				targetDir,
			)

			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return err
			}
			namespace.Bind(src, dst, namespace.MREPL)
		}
	}
	return nil
}
func (b *Builder) prepare(n *graph.Node) error {
	// prepare
	if err := namespace.New(b.buildpath(n)); err != nil {
		return err
	}
	if err := b.bindChildOutputs(n); err != nil {
		return err
	}

	return b.createOutputDirs(n)

}
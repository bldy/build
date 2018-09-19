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

func (b *Builder) build(e *executor.Executor, n *graph.Node) error {
	n.Start = time.Now().UnixNano()
	n.Status = build.Fail
	err := n.Target.Build(e)
	if err == nil {
		n.Status = build.Success
	}

	n.End = time.Now().UnixNano()
	n.Output = e.CombinedLog()

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

		finish := func(err error) {
			b.notifier.Update(job)
			if err != nil {
				b.notifier.Error(err)
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

		if b.cached(job) {
			finish(b.builderror(job))
		} else {
			if err := ctx.Err(); err != nil {
				finish(err)
				return
			}

			ns, err := b.prepare(ctx, job)
			if err != nil {
				finish(err)
				return
			}
			e := executor.New(ctx, ns)
			finish(b.build(e, job))
		}

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

func (b *Builder) bindChildOutputs(n *graph.Node, ns namespace.Namespace) error {
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
			ns.Bind(dst, src, namespace.MREPL)
		}
	}
	return nil
}
func (b *Builder) prepare(ctx context.Context, n *graph.Node) (namespace.Namespace, error) {
	// prepare
	ns, err := b.newnamespace(n)
	if err != nil {
		return nil, err
	}
	if ws, ok := ns.(namespace.Workspace); ok {
		ws.MountWorkspace(n.Target.Workspace().AbsPath())
	}
	if err := b.bindChildOutputs(n, ns); err != nil {
		return nil, err
	}

	if err := b.createOutputDirs(n); err != nil {
		return nil, err
	}

	return ns, nil
}
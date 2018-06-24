// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
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
}

type Notifier interface {
	Update(*graph.Node)
	Error(error)
	Done(time.Duration)
}

type Config struct {
	UseCache bool
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
	if c.Cache == nil {
		c.Cache = bldyCache()
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
	b.visit(b.graph.Root)
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
	return fmt.Errorf("%s: \n%s", err, n.Output)
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

		go b.notifier.Update(job)

		if b.cached(job) {
			if err := b.builderror(job); err != nil {
				go b.notifier.Error(err)
			}
			go b.notifier.Update(job)
			continue
		}

		// prepare
		if err := namespace.New(b.buildpath(job)); err != nil {
			go b.notifier.Error(errors.Wrap(err, "build"))
		}

		if err := ctx.Err(); err != nil {
			go b.notifier.Error(err)
			continue
		}

		if err := b.build(ctx, job); err != nil {
			go b.notifier.Error(err)
		}

		job.Once.Do(func() {
			for _, parent := range job.Parents {
				parent.WG.Done()
			}
		})

		go b.notifier.Update(job)

		if job.IsRoot {
			install(job, *b.config.BuildOut)
			b.notifier.Done(time.Now().Sub(b.start))
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

func install(job *graph.Node, buildOut string) error {
	if err := os.MkdirAll(
		buildOut,
		os.ModeDir|os.ModePerm,
	); err != nil {
		l.Fatalf("copying job %s failed: %s", job.Target.GetName(), err.Error())
	}

	for dst, src := range job.Target.Installs() {
		target := filepath.Base(dst)
		targetDir := strings.TrimRight(dst, target)

		buildOutTarget := filepath.Join(
			buildOut,
			targetDir,
		)
		log.Println(dst)

		if err := os.MkdirAll(
			buildOutTarget,
			os.ModeDir|os.ModePerm,
		); err != nil {
			l.Fatalf("linking job %s failed: %s", job.Target.GetName(), err.Error())
		}
		srcp, _ := filepath.EvalSymlinks(
			filepath.Join(
				buildOut,
				fmt.Sprintf("%s-%x", job.Target.GetName(), job.HashNode()),
				src,
			),
		)

		dstp := filepath.Join(
			buildOutTarget,
			target,
		)

		in, err := os.Open(srcp)
		if err != nil {
			l.Fatalf("copy: can't finiliaze %s. copying %q to %q failed: %s\n", job.Target.GetName(), srcp, dstp, err)
		}
		out, err := os.Create(dstp)
		if err != nil {
			l.Fatal(err)
		}

		if _, err := io.Copy(out, in); err != nil {
			l.Fatalf("copy: can't finiliaze %s. copying from %q to %q failed: %s\n", job.Target.GetName(), src, dst, err)
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

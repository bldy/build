// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"time"

	"io/ioutil"

	"strings"

	"bldy.build/build"
	"bldy.build/build/graph"
	"bldy.build/build/project"
)

const (
	SCSSLOG = "success"
	FAILLOG = "fail"
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
	Done        chan *graph.Node
	Error       chan error
	Timeout     chan bool
	Updates     chan *graph.Node
	ptr         *graph.Node
	graph       *graph.Graph
	pq          *p
}

func New(g *graph.Graph) (c Builder) {
	c.Error = make(chan error)
	c.Done = make(chan *graph.Node)
	c.Updates = make(chan *graph.Node)
	var err error
	c.Wd, err = os.Getwd()
	if err != nil {
		l.Fatal(err)
	}
	c.pq = newP()
	c.graph = g
	c.ProjectPath = project.Root()
	return
}

var (
	BLDYCACHE = bldyCache()
	l         = log.New(os.Stdout, "builder: ", 0)
)

func bldyCache() string {
	usr, err := user.Current()
	if err != nil {
		l.Fatal(err)
	}
	return path.Join(usr.HomeDir, "/.cache/bldy")
}

func (b *Builder) Execute(ctx context.Context, r int) {

	for i := 0; i < r; i++ {
		go b.work(ctx, i)
	}

	if b.graph == nil {
		l.Fatal("Couldn't find the build graph")
	}
	b.visit(b.graph.Root)
}

func (b *Builder) build(ctx context.Context, n *graph.Node) (err error) {
	buildErr := ctx.Err()
	if err != nil {
		return err
	}
	nodeHash := fmt.Sprintf("%s-%x", n.Target.GetName(), n.HashNode())
	outDir := filepath.Join(
		BLDYCACHE,
		nodeHash,
	)
	// check if this node was build before
	if _, err := os.Lstat(outDir); !os.IsNotExist(err) {
		n.Cached = true
		if file, err := os.Open(filepath.Join(outDir, FAILLOG)); err == nil {
			errString, _ := ioutil.ReadAll(file)
			return fmt.Errorf("%s", errString)
		} else if _, err := os.Lstat(filepath.Join(outDir, SCSSLOG)); err == nil {
			return nil
		}
	}

	os.MkdirAll(outDir, os.ModeDir|os.ModePerm)

	// check failed builds.
	for _, e := range n.Children {
		if e.Status == build.Fail {
			buildErr = fmt.Errorf("dependency %s failed to build", e.Target.GetName())
		} else {
			for dst, src := range e.Target.Installs() {

				target := filepath.Base(dst)
				targetDir := strings.TrimRight(dst, target)

				if targetDir != "" {
					if err := os.MkdirAll(
						filepath.Join(
							outDir,
							targetDir,
						),
						os.ModeDir|os.ModePerm,
					); err != nil {
						l.Fatalf("installing dependency %s for %s: %s", e.Target.GetName(), n.Target.GetName(), err.Error())
					}
				}
				os.Symlink(
					filepath.Join(
						BLDYCACHE,
						fmt.Sprintf("%s-%x", e.Target.GetName(), e.HashNode()),
						src,
					),
					filepath.Join(
						outDir,
						targetDir,
						target),
				)

			}
		}
	}

	runner := build.NewRunner(ctx, outDir)
	n.Start = time.Now().UnixNano()

	buildErr = n.Target.Build(runner)
	n.End = time.Now().UnixNano()

	logName := FAILLOG
	if buildErr == nil {
		logName = SCSSLOG
	}
	if logFile, err := os.Create(filepath.Join(outDir, logName)); err != nil {
		l.Fatalf("error creating log for %s: %s", n.Target.GetName(), err.Error())
	} else {
		log := runner.Log()
		buf := bytes.Buffer{}
		for _, logEntry := range log {
			buf.WriteString(logEntry.String())
		}
		n.Output = buf.String()
		_, err = logFile.Write(buf.Bytes())
		if err != nil {
			l.Fatalf("error writing log for %s: %s", n.Target.GetName(), err.Error())
		}
		if buildErr != nil {
			return fmt.Errorf("%s: \n%s", buildErr, buf.Bytes())
		}
	}

	return buildErr
}

func (b *Builder) work(ctx context.Context, workerNumber int) {

	for {
		job := b.pq.pop()
		job.Worker = fmt.Sprintf("%d", workerNumber)
		if job.Status != build.Pending {
			continue
		}
		job.Lock()
		defer job.Unlock()

		job.Status = build.Building

		b.Updates <- job
		buildErr := b.build(ctx, job)

		if buildErr != nil {
			job.Status = build.Fail
			b.Updates <- job
			b.Error <- buildErr

		} else {
			job.Status = build.Success

			b.Updates <- job
		}

		if !job.IsRoot {
			b.Done <- job
			job.Once.Do(func() {
				for _, parent := range job.Parents {
					parent.WG.Done()
				}
			})
		} else {
			install(job)

			b.Done <- job
			close(b.Done)
			return
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
	n.CountDependents()
	b.pq.push(n)
}

func install(job *graph.Node) error {
	buildOut := project.BuildOut()
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
		if err := os.MkdirAll(
			buildOutTarget,
			os.ModeDir|os.ModePerm,
		); err != nil {
			l.Fatalf("linking job %s failed: %s", job.Target.GetName(), err.Error())
		}
		srcp, _ := filepath.EvalSymlinks(
			filepath.Join(
				BLDYCACHE,
				fmt.Sprintf("%s-%x", job.Target.GetName(), job.HashNode()),
				src,
			))

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

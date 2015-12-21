// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context // import "sevki.org/build/context"

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"io/ioutil"

	"sevki.org/build/ast"
)

func (c *Context) Execute(d time.Duration, r int) {

	for i := 0; i < r; i++ {

		go c.work(c.BuildQueue, i)
		c.Updates <- Update{
			Worker:    i,
			TimeStamp: time.Now(),
			Status:    Pending,
		}
	}

	go func() {
		if d > 0 {
			time.Sleep(d)
			c.Timeout <- true
		}
	}()
	if c.Root == nil {
		log.Fatal("Couldn't find the root node.")
	}
	c.visit(c.Root)
}
func build(n *Node) (err error) {
	var buildErr error

	// check failed builds.
	for _, e := range n.Edges {
		if e.Status == Fail {
			buildErr = fmt.Errorf("dependency %s failed to build", e.Target.GetName())
		}
	}

	nodeHash := fmt.Sprintf("%s-%x", n.Target.GetName(), n.hashNode())

	outDir := filepath.Join(
		"/tmp",
		"build",
		nodeHash,
	)

	// check if this node was build before
	if _, err := os.Lstat(outDir); !os.IsNotExist(err) {
		if file, err := os.Open(filepath.Join(outDir, "failed")); err == nil {
			errString, _ := ioutil.ReadAll(file)
			return fmt.Errorf("%s", errString)
		} else if _, err := os.Lstat(filepath.Join(outDir, "success")); err == nil {
			if err := n.Target.Install(); err != nil {
				return err
			}
			return nil
		}
	}

	os.MkdirAll(outDir, os.ModeDir|os.ModePerm)

	os.Chdir(outDir)

	buildErr = n.Target.Build()
	logName := "failed"
	if buildErr == nil {
		logName = "success"

		n.Target.Install()
	}
	if logFile, err := os.Create(filepath.Join(outDir, logName)); err != nil {
		log.Fatal(err)
	} else {
		_, err := io.Copy(logFile, n.Target.Reader())
		if err != nil {
			log.Fatal(err)
		}
	}

	return buildErr
}

func (c *Context) work(jq chan *Node, workerNumber int) {

	for {
		select {
		case job := <-jq:
			c.Updates <- Update{
				Worker:    workerNumber,
				TimeStamp: time.Now(),
				Target:    job.Target.GetName(),
				Status:    Started,
			}

			buildErr := build(job)

			if buildErr != nil {
				c.Updates <- Update{
					Worker:    workerNumber,
					TimeStamp: time.Now(),
					Target:    job.Target.GetName(),
					Status:    Fail,
				}

				c.Error <- buildErr
				job.Status = Fail
			} else {
				c.Updates <- Update{
					Worker:    workerNumber,
					TimeStamp: time.Now(),
					Target:    job.Target.GetName(),
					Status:    Success,
				}

				job.Status = Success
			}

			c.Done <- job.Target

			if job.parentWg != nil {
				job.parentWg.Done()
			} else {
				close(c.Done)

				return
			}

		}
	}

}

type Edges map[string]*Node

type STATUS int

const (
	Success STATUS = iota
	Fail
	Pending
	Started
	Fatal
)

type Node struct {
	Target   ast.Target
	Edges    Edges
	wg       sync.WaitGroup
	parentWg *sync.WaitGroup
	Status   STATUS
	Output   string
}

func (n *Node) hashNode() []byte {
	h := sha1.New()
	h.Write(n.Target.Hash())
	for _, e := range n.Edges {
		h.Write(e.hashNode())
	}
	return h.Sum(nil)
}

func (c *Context) visit(n *Node) {

	// This is not an airplane so let's make sure children get their masks on before the parents.
	for _, child := range n.Edges {
		// Make sure we block this routine until all the children are done
		n.wg.Add(1)

		// Visit children first
		go c.visit(child)
	}

	n.wg.Wait()

	c.BuildQueue <- n
}

// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder // import "sevki.org/build/builder"

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"io/ioutil"

	"strings"

	"sevki.org/build/build"
)

const (
	SCSSLOG = "success"
	FAILLOG = "fail"
)

func (b *Builder) Execute(d time.Duration, r int) {

	for i := 0; i < r; i++ {

		go b.work(b.BuildQueue, i)
		b.Updates <- Update{
			Worker:    i,
			TimeStamp: time.Now(),
			Status:    Pending,
		}
	}

	go func() {
		if d > 0 {
			time.Sleep(d)
			b.Timeout <- true
		}
	}()
	if b.Root == nil {
		log.Fatal("Couldn't find the root node.")
	}
	b.visit(b.Root)
}
func (b *Builder) build(n *Node) (err error) {

	var buildErr error

	nodeHash := fmt.Sprintf("%s-%x", n.Target.GetName(), n.hashNode())

	outDir := filepath.Join(
		"/tmp",
		"build",
		nodeHash,
	)
	// check if this node was build before
	if _, err := os.Lstat(outDir); !os.IsNotExist(err) {
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
		if e.Status == Fail {
			buildErr = fmt.Errorf("dependency %s failed to build", e.Target.GetName())
		} else {
			for file, folder := range e.Target.Installs() {
				if folder != "" {
					if err := os.MkdirAll(
						filepath.Join(
							outDir,
							folder,
						),
						os.ModeDir|os.ModePerm,
					); err != nil {
						log.Fatalf("installing dependency %s for %s: %s", e.Target.GetName(), n.Target.GetName(), err.Error())
					}
				}
				os.Symlink(
					filepath.Join(
						"/tmp",
						"build",
						fmt.Sprintf("%s-%x", e.Target.GetName(), e.hashNode()),
						file,
					),
					filepath.Join(
						outDir,
						folder,
						file),
				)

			}
		}
	}

	context := build.NewContext(outDir)
	buildErr = n.Target.Build(context)

	logName := FAILLOG
	if buildErr == nil {
		logName = SCSSLOG
	}
	if logFile, err := os.Create(filepath.Join(outDir, logName)); err != nil {
		log.Fatalf("error creating log for %s:", n.Target.GetName(), err.Error())
	} else {
		_, err := io.Copy(logFile, context.Stdout())
		if err != nil {
			log.Fatalf("error writing log for %s:", n.Target.GetName(), err.Error())
		}
	}

	return buildErr
}

func (b *Builder) work(jq chan *Node, workerNumber int) {

	for {
		select {
		case job := <-jq:
			if job.Status != Pending {
				continue
			}
			job.Lock()
			defer job.Unlock()
			job.Status = Building

			b.Updates <- Update{
				Worker:    workerNumber,
				TimeStamp: time.Now(),
				Target:    job.Target.GetName(),
				Status:    Started,
			}
			buildErr := b.build(job)

			if buildErr != nil {
				job.Status = Fail
				b.Updates <- Update{
					Worker:    workerNumber,
					TimeStamp: time.Now(),
					Target:    job.Target.GetName(),
					Status:    Fail,
				}

				b.Error <- buildErr

			} else {
				job.Status = Success

				b.Updates <- Update{
					Worker:    workerNumber,
					TimeStamp: time.Now(),
					Target:    job.Target.GetName(),
					Status:    Success,
				}
			}

			b.Done <- job.Target
			if len(job.Parents) > 0 {

				job.once.Do(func() {
					for _, parent := range job.Parents {
						parent.wg.Done()
					}
				})

			} else {
				close(b.Done)

				return
			}

		}
	}

}

type STATUS int

const (
	Success STATUS = iota
	Fail
	Pending
	Started
	Fatal
	Building
)

type ByName []*Node

func (a ByName) Len() int      { return len(a) }
func (a ByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool {
	return strings.Compare(a[i].Target.GetName(), a[j].Target.GetName()) > 0
}

func (n *Node) hashNode() []byte {
	h := sha1.New()
	h.Write(n.Target.Hash())
	var bn ByName
	for _, e := range n.Children {
		bn = append(bn, e)

	}
	sort.Sort(bn)
	for _, e := range bn {
		h.Write(e.hashNode())
	}
	return h.Sum(nil)
}

func (b *Builder) visit(n *Node) {

	// This is not an airplane so let's make sure children get their masks on before the parents.
	for _, child := range n.Children {
		// Visit children first
		go b.visit(child)
	}

	n.wg.Wait()

	b.BuildQueue <- n
}

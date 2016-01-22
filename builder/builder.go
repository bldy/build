// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package builder parses build graphs and coordinates builds
package builder // import "sevki.org/build/builder"

import (
	"log"
	"os"
	"time"

	"sync"

	"sevki.org/build"
	"sevki.org/build/parser"
	"sevki.org/build/postprocessor"
	"sevki.org/build/preprocessor"
	"sevki.org/build/processor"
	"sevki.org/build/util"
)

type Update struct {
	TimeStamp time.Time
	Target    string
	Status    STATUS
	Worker    int
}
type Builder struct {
	Origin      string
	Wd          string
	ProjectPath string
	Nodes       map[string]*Node
	Total       int
	Done        chan *Node
	Error       chan error
	Timeout     chan bool
	Updates     chan Update
	Root, ptr   *Node
	BuildQueue  chan *Node
}

func New() (c Builder) {
	c.Nodes = make(map[string]*Node)
	c.Error = make(chan error)
	c.Done = make(chan *Node)
	c.BuildQueue = make(chan *Node)
	c.Updates = make(chan Update)
	var err error
	c.Wd, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	c.ProjectPath = util.GetProjectPath()
	return
}

type WorkGroupPublic struct {
	Waiters int
	wg      sync.WaitGroup
	sync.Mutex
}

func (wg *WorkGroupPublic) Add(i int) {
	wg.Lock()
	defer wg.Unlock()
	wg.Waiters += i
	wg.wg.Add(i)

}

func (wg *WorkGroupPublic) Wait() {
	wg.wg.Wait()
}
func (wg *WorkGroupPublic) Done() {
	wg.Lock()
	defer wg.Unlock()
	wg.Waiters = wg.Waiters - 1
}

type Node struct {
	IsRoot   bool
	Target   build.Target
	Children map[string]*Node
	Parents  map[string]*Node `json:"-"`
	Url      parser.TargetURL
	wg       sync.WaitGroup
	Status   STATUS
	Output   string
	once     sync.Once
	sync.Mutex
}

func (b *Builder) getTarget(url parser.TargetURL) (n *Node) {

	if gnode, ok := b.Nodes[url.String()]; ok {
		return gnode
	} else {

		doc, err := parser.ReadBuildFile(url, b.Wd)

		if err != nil {
			log.Fatalf("getting target %s failed :%s", url.String(), err.Error())
		}

		if err := preprocessor.Process(doc); err != nil {
			log.Fatalf("error processing document: %s", err.Error())
		}

		var p processor.Processor

		for name, t := range p.Process(doc) {
			if t.GetName() != url.Target {
				continue
			}
			xu := parser.TargetURL{
				Package: url.Package,
				Target:  name,
			}

			node := Node{
				Target:   t,
				Children: make(map[string]*Node),
				Parents:  make(map[string]*Node),
				once:     sync.Once{},
				wg:       sync.WaitGroup{},
				Status:   Pending,
				Url:      xu,
			}

			post := postprocessor.New(url.Package)

			err := post.ProcessDependencies(node.Target)
			if err != nil {
				log.Fatal(err)
			}

			var deps []build.Target

			for _, d := range node.Target.GetDependencies() {
				c := b.Add(d)
				node.wg.Add(1)

				deps = append(deps, c.Target)

				node.Children[d] = c
				c.Parents[xu.String()] = &node

			}

			if err := post.ProcessPaths(t, deps); err != nil {
				log.Fatalf("path processing: %s", err.Error())
			}

			b.Nodes[xu.String()] = &node
			if t.GetName() == url.Target {
				n = &node
			}

		}

		if n == nil {
			log.Fatalf("we couldn't find %s, %s", url.String())
		}
		return n
	}

}

func (b *Builder) Add(t string) *Node {
	return b.getTarget(parser.NewTargetURLFromString(t))
}

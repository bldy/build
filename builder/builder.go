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
	Done        chan build.Target
	Error       chan error
	Timeout     chan bool
	Updates     chan Update
	Root, ptr   *Node
	BuildQueue  chan *Node
}

func New() (c Builder) {
	c.Nodes = make(map[string]*Node)
	c.Error = make(chan error)
	c.Done = make(chan build.Target)
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

type Edges map[string]*Node

type Node struct {
	Target   build.Target
	Children Edges
	Parents  Edges `json:"-"`
	wg       sync.WaitGroup
	Status   STATUS
	Output   string
	once     sync.Once
	sync.RWMutex
}

func (b *Builder) getTarget(url parser.TargetURL) (n *Node) {

	if gnode, ok := b.Nodes[url.String()]; ok {
		return gnode
	} else {

		doc, err := parser.ReadBuildFile(url, b.Wd)
		if err != nil {
			log.Fatalf("getting target %s failed :%s", url.String(), err.Error())
		}

		var pp parser.PreProcessor

		for name, t := range pp.PreProcess(doc) {
			xu := parser.TargetURL{
				Package: url.Package,
				Target:  name,
			}

			node := Node{
				Target:   t,
				Children: make(Edges),
				Parents:  make(Edges),
				once:     sync.Once{},
				wg:       sync.WaitGroup{},
				Status:   Pending,
			}

			post := postprocessor.New(url.Package)

			post.ProcessDependencies(node.Target)

			node.wg.Add(len(t.GetDependencies()))
			b.Total += len(t.GetDependencies())
			var deps []build.Target

			for _, d := range node.Target.GetDependencies() {
				c := b.Add(d)
				deps = append(deps, c.Target)
				node.Children[c.Target.GetName()] = c
				c.Parents[node.Target.GetName()] = &node

			}

			post.ProcessPaths(t, deps)

			b.Nodes[xu.String()] = &node
			if t.GetName() == url.Target {
				n = &node
			}

		}
		if n == nil {
			log.Fatalf("we couldn't find %s", url.String())
		}
		return n
	}

}

func (b *Builder) Add(t string) *Node {
	return b.getTarget(parser.NewTargetURLFromString(t))
}

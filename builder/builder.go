// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package context defines and helps create a context for the build graph
package builder // import "sevki.org/build/builder"

import (
	"fmt"
	"log"
	"os"
	"time"

	"sync"

	"sevki.org/build/build"
	"sevki.org/build/parser"
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
	Targets     map[string]build.Target
	Total       int
	Done        chan build.Target
	Error       chan error
	Timeout     chan bool
	Updates     chan Update
	Root, ptr   *Node
	BuildQueue  chan *Node
}

func (c *Builder) getTarget(name string) build.Target {
	url := parser.NewTargetURLFromString(name)

	if t, ok := c.Targets[url.String()]; ok {
		return t
	} else {

		doc, err := parser.ReadBuildFile(url, c.Wd)
		if err != nil {
			log.Fatalf("getting target %s failed :%s", name, err.Error())
		}

		var x build.Target
		var pp parser.PreProcessor

		for name, t := range pp.Process(doc) {
			xu := parser.TargetURL{
				Package: url.Package,
				Target:  name,
			}

			c.Targets[xu.String()] = t
			if t.GetName() == url.Target {
				x = t
			}
		}

		if x == nil {
			log.Fatal("we couldn't find the your dependency")
		}

		return x

	}

}
func (c *Builder) Parse(t string) {
	if x := c.getTarget(t); x != nil {
		c.add(x)
	}
}

func New() (c Builder) {
	c.Targets = make(map[string]build.Target)
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

func (c *Builder) add(t build.Target) {
	c.Total++
	curNode := Node{
		Target: t,
		Edges:  make(Edges),
		wg:     sync.WaitGroup{},
		Status: Pending,
	}
	if c.Root == nil {
		c.Root = &curNode
	}

	if c.ptr != nil {
		edgeName := fmt.Sprintf("%s:%s", c.ptr.Target.GetName(), t.GetName())
		log.Println(edgeName)
		c.ptr.Edges[edgeName] = &curNode
		curNode.parentWg = &c.ptr.wg
	}

	tmp := c.ptr
	c.ptr = &curNode

	for _, d := range t.GetDependencies() {
		c.Parse(d)
	}

	c.ptr = tmp
}

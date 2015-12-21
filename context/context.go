// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package context defines and helps create a context for the build graph
package context // import "sevki.org/build/context"

import (
	"fmt"
	"log"
	"os"
	"time"

	"sync"

	"sevki.org/build/ast"
	"sevki.org/build/parser"
	"sevki.org/build/util"
)

type Update struct {
	TimeStamp time.Time
	Target    string
	Status    STATUS
	Worker    int
}
type Context struct {
	Origin      string
	Wd          string
	ProjectPath string
	Targets     map[string]ast.Target
	Total       int
	Done        chan ast.Target
	Error       chan error
	Timeout     chan bool
	Updates     chan Update
	Root, ptr   *Node
	BuildQueue  chan *Node
}

func (c *Context) getTarget(name string) ast.Target {

	if t, ok := c.Targets[name]; ok {
		return t
	} else {
		url := parser.NewTargetURLFromString(name)
		doc, err := parser.ReadBuildFile(url, c.Wd)
		if err != nil {
			log.Fatalf("getting target %s failed:%s", name, err.Error())
		}

		var x ast.Target
		var pp parser.PreProcessor

		for name, t := range pp.Process(doc) {
			c.Targets[name] = t
			if t.GetName() == url.Target {
				x = t
			}
		}
		return x

	}

}
func (c *Context) Parse(t string) {
	if x := c.getTarget(t); x != nil {
		c.add(x)
	}
}

func New() (c Context) {
	c.Targets = make(map[string]ast.Target)
	c.Error = make(chan error)
	c.Done = make(chan ast.Target)
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

func (c *Context) add(t ast.Target) {
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

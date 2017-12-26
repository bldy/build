// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package graph parses and generates build graphs
package graph

import (
	"fmt"
	"log"
	"os"

	"sync"

	"bldy.build/build"
	"bldy.build/build/skylark"

	"bldy.build/build/postprocessor"
	bldytrg "bldy.build/build/rules/build"
	"bldy.build/build/url"
)

var (
	l = log.New(os.Stdout, "graph: ", 0)
)

// Node encapsulates a target and represents a node in the build graph.
type Node struct {
	IsRoot        bool       `json:"-"`
	Target        build.Rule `json:"-"`
	Type          string
	Parents       map[string]*Node `json:"-"`
	URL           url.URL
	Worker        string
	PriorityCount int
	WG            sync.WaitGroup
	Status        build.Status
	Cached        bool
	Start, End    int64
	Hash          string
	Output        string `json:"-"`
	Once          sync.Once
	sync.Mutex
	Children map[string]*Node
	hash     []byte
}

// Graph represents a build graph
type Graph struct {
	Root  *Node
	vm    build.VM
	Nodes map[string]*Node
}

// New returns a new build graph relatvie to the working directory
func New(wd, target string) *Graph {
	vm, err := skylark.New(wd)
	if err != nil {
		return nil
	}
	g := Graph{
		vm:    vm,
		Nodes: make(map[string]*Node),
	}
	g.Root = g.getTarget(url.Parse(target))
	g.Root.IsRoot = true
	return &g
}

// CountDependents counts how many nodes directly and indirectly depend on
// this node
func (n *Node) Priority() int {
	if n.PriorityCount < 0 {
		p := 0
		for _, c := range n.Parents {
			p += c.Priority() + 1
		}
		n.PriorityCount = p
	}
	return n.PriorityCount
}

func (g *Graph) getTarget(u url.URL) (n *Node) {
	if gnode, ok := g.Nodes[u.String()]; ok {
		return gnode
	}
	t, err := g.vm.GetTarget(u)
	if err != nil {
		log.Fatal(err)
	}
	xu := url.URL{
		Package: u.Package,
		Target:  t.GetName(),
	}

	node := Node{
		Target:        t,
		Type:          fmt.Sprintf("%T", t)[1:],
		Children:      make(map[string]*Node),
		Parents:       make(map[string]*Node),
		Once:          sync.Once{},
		WG:            sync.WaitGroup{},
		Status:        build.Pending,
		URL:           xu,
		PriorityCount: -1,
	}

	post := postprocessor.New(u.Package)

	err = post.ProcessDependencies(node.Target)
	if err != nil {
		l.Fatal(err)
	}

	var deps []build.Rule

	//group is a special case
	var group *bldytrg.Group
	switch node.Target.(type) {
	case *bldytrg.Group:
		group = node.Target.(*bldytrg.Group)
		group.Exports = make(map[string]string)
	}
	for _, d := range node.Target.GetDependencies() {
		c := g.getTarget(url.Parse(d))
		node.WG.Add(1)
		if group != nil {
			for dst := range c.Target.Installs() {
				group.Exports[dst] = dst
			}
		}
		deps = append(deps, c.Target)

		node.Children[d] = c
		c.Parents[xu.String()] = &node
	}

	if err := post.ProcessPaths(t, deps); err != nil {
		l.Fatalf("path processing: %s", err.Error())
	}

	g.Nodes[xu.String()] = &node
	if t.GetName() == u.Target {
		n = &node
	} else {
		l.Fatalf("target name %q and url target %q don't match", t.GetName(), u.Target)
	}
	return n
}

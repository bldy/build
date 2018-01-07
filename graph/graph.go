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
	"bldy.build/build/label"
	"bldy.build/build/skylark"
	"github.com/pkg/errors"

	"bldy.build/build/postprocessor"
	bldytrg "bldy.build/build/rules/build"
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
	Label         label.Label
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
func New(wd, target string) (*Graph, error) {
	vm, err := skylark.New(wd)
	if err != nil {
		return nil, errors.Wrap(err, "new graph")
	}
	g := Graph{
		vm:    vm,
		Nodes: make(map[string]*Node),
	}
	label, err := label.Parse(target)
	if err != nil {
		return nil, errors.Wrap(err, "new graph")
	}
	g.Root = g.getTarget(label)
	g.Root.IsRoot = true
	return &g, nil
}

// Priority counts how many nodes directly and indirectly depend on
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

func (g *Graph) getTarget(lbl *label.Label) (n *Node) {
	if gnode, ok := g.Nodes[lbl.String()]; ok {
		return gnode
	}
	t, err := g.vm.GetTarget(lbl)
	if err != nil {
		l.Fatal(err)
	}
	nLbl := label.Label{
		Package: lbl.Package,
		Name:    t.GetName(),
	}

	node := Node{
		Target:        t,
		Type:          fmt.Sprintf("%T", t)[1:],
		Children:      make(map[string]*Node),
		Parents:       make(map[string]*Node),
		Once:          sync.Once{},
		WG:            sync.WaitGroup{},
		Status:        build.Pending,
		Label:         nLbl,
		PriorityCount: -1,
	}

	post := postprocessor.New(nLbl.Package)

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
		depLbl, err := label.Parse(d)
		c := g.getTarget(depLbl)
		if err != nil {
			l.Printf("%q is not a valid url", depLbl)
			continue
		}
		node.WG.Add(1)
		if group != nil {
			for dst := range c.Target.Installs() {
				group.Exports[dst] = dst
			}
		}
		deps = append(deps, c.Target)

		node.Children[d] = c
		c.Parents[nLbl.String()] = &node
	}

	if err := post.ProcessPaths(t, deps); err != nil {
		l.Fatalf("path processing: %s", err.Error())
	}

	g.Nodes[nLbl.String()] = &node
	if t.GetName() == lbl.Name {
		n = &node
	} else {
		l.Fatalf("target name %q and url target %q don't match", t.GetName(), lbl.Name)
	}
	return n
}

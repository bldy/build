// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package graph parses and generates build graphs
package graph

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"sync"

	"bldy.build/build"
	"bldy.build/build/blaze"
	"bldy.build/build/blaze/postprocessor"
	"bldy.build/build/racy"
	bldytrg "bldy.build/build/targets/build"
	"bldy.build/build/url"
)

var (
	l = log.New(os.Stdout, "graph: ", 0)
)

// Node encapsulates a target and represents a node in the build graph.
type Node struct {
	IsRoot     bool         `json:"-"`
	Target     build.Target `json:"-"`
	Type       string
	Parents    map[string]*Node `json:"-"`
	URL        url.URL
	Worker     string
	Priority   int
	WG         sync.WaitGroup
	Status     build.Status
	Cached     bool
	Start, End int64
	Hash       string
	Output     string `json:"-"`
	Once       sync.Once
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
	g := Graph{
		vm:    blaze.NewVM(wd),
		Nodes: make(map[string]*Node),
	}
	g.Root = g.getTarget(url.Parse(target))
	g.Root.IsRoot = true
	return &g
}

// CountDependents counts how many nodes directly and indirectly depend on
// this node
func (n *Node) CountDependents() int {
	if n.Priority < 0 {
		p := 0
		for _, c := range n.Parents {
			p += c.CountDependents() + 1
		}
		n.Priority = p
	}
	return n.Priority
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
		Target:   t,
		Type:     fmt.Sprintf("%T", t)[1:],
		Children: make(map[string]*Node),
		Parents:  make(map[string]*Node),
		Once:     sync.Once{},
		WG:       sync.WaitGroup{},
		Status:   build.Pending,
		URL:      xu,
		Priority: -1,
	}

	post := postprocessor.New(u.Package)

	err = post.ProcessDependencies(node.Target)
	if err != nil {
		l.Fatal(err)
	}

	var deps []build.Target

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

// HashNode calculates the hash of a node
func (n *Node) HashNode() []byte {

	// node hashes should not change after a build,
	// they should be deterministic, therefore they can and should be cached.
	if len(n.hash) > 0 {
		return n.hash
	}
	n.hash = n.Target.Hash()
	var bn ByName
	for _, e := range n.Children {
		bn = append(bn, e)
	}
	sort.Sort(bn)
	for _, e := range bn {
		n.hash = racy.XOR(e.HashNode(), n.hash)
	}
	n.Hash = fmt.Sprintf("%x", n.hash)
	return n.hash
}

// ByName sorts dependencies by name so we can have reproduceable builds.
type ByName []*Node

func (a ByName) Len() int      { return len(a) }
func (a ByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool {
	return strings.Compare(a[i].Target.GetName(), a[j].Target.GetName()) > 0
}

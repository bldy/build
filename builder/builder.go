// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package builder parses build graphs and coordinates builds
package builder // import "sevki.org/build/builder"

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"sync"

	"sevki.org/build"
	"sevki.org/build/parser"
	"sevki.org/build/postprocessor"
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

type Node struct {
	IsRoot     bool         `json:"-"`
	Target     build.Target `json:"-"`
	Type       string
	Parents    map[string]*Node `json:"-"`
	Url        parser.TargetURL
	wg         sync.WaitGroup
	Status     STATUS
	Start, End int64
	Hash       string
	Output     string `json:"-"`
	once       sync.Once
	sync.Mutex
	Children map[string]*Node
	hash     []byte
}

func (b *Builder) getTarget(url parser.TargetURL) (n *Node) {

	if gnode, ok := b.Nodes[url.String()]; ok {
		return gnode
	} else {
		p, err := processor.NewProcessorFromURL(url, b.Wd)
		if err != nil {
			log.Fatal(err)
		}
		go p.Run()
		// bug(sevki): this is a really bad way of doing this, there should me
		// some caching mechanism for this, it is yet to come !!
		for t := <-p.Targets; t != nil; t = <-p.Targets {
			if t.GetName() != url.Target {
				continue
			}
			xu := parser.TargetURL{
				Package: url.Package,
				Target:  t.GetName(),
			}

			node := Node{
				Target:   t,
				Type:     fmt.Sprintf("%T", t)[1:],
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
			log.Fatalf("we couldn't find target %s", url.String())
		}
		return n
	}

}

func (b *Builder) Add(t string) *Node {
	return b.getTarget(parser.NewTargetURLFromString(t))
}

func (n *Node) HashNode() []byte {
	// node hashes should not change after a build,
	// they should be deterministic, therefore they should and can be cached.
	if len(n.hash) > 0 {
		return n.hash
	}
	h := sha1.New()
	h.Write(n.Target.Hash())
	util.HashStrings(h, n.Target.GetDependencies())
	var bn ByName
	for _, e := range n.Children {
		bn = append(bn, e)
	}
	sort.Sort(bn)
	for _, e := range bn {
		h.Write(e.HashNode())
	}
	n.hash = h.Sum(nil)
	n.Hash = fmt.Sprintf("%x",n.hash)
	return n.hash
}

type ByName []*Node

func (a ByName) Len() int      { return len(a) }
func (a ByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool {
	return strings.Compare(a[i].Target.GetName(), a[j].Target.GetName()) > 0
}

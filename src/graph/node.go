package graph

import (
	"fmt"
	"sync"

	"bldy.build/bldy/src/build"
	"bldy.build/bldy/src/url"
)

// NewNode takes a label and a rule and returns it as a Graph Node
func NewNode(u *url.URL, t build.Rule) Node {
	return Node{
		Target:      t,
		Type:        fmt.Sprintf("%T", t)[1:],
		Children:    make(map[string]*Node),
		Parents:     make(map[string]*Node),
		Once:        sync.Once{},
		WG:          sync.WaitGroup{},
		Status:      build.Pending,
		url:         u,
		ndependents: -1,
	}
}

// Node encapsulates a target and represents a node in the build graph.
type Node struct {
	ID     string
	url    *url.URL
	Type   string
	Worker string

	WG         sync.WaitGroup
	Status     build.Status
	Cached     bool
	Start, End int64
	Hash       string

	hash        []byte
	ndependents int

	Once sync.Once
	sync.Mutex
	Children map[string]*Node

	Workspace string

	Output  string           `json:"-"`
	IsRoot  bool             `json:"-"`
	Target  build.Task       `json:"-"`
	Parents map[string]*Node `json:"-"`
}

// Priority counts how many nodes directly and indirectly depend on
// this node
func (n *Node) Priority() int {
	if n.ndependents < 0 {
		p := 0
		for _, c := range n.Parents {
			p += c.Priority() + 1
		}
		n.ndependents = p
	}
	return n.ndependents
}

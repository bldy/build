package graph

import (
	"fmt"
	"sync"

	"bldy.build/build"
	"bldy.build/build/label"
)

// NewNode takes a label and a rule and returns it as a Graph Node
func NewNode(l label.Label, t build.Rule) Node {
	return Node{
		Target:        t,
		Type:          fmt.Sprintf("%T", t)[1:],
		Children:      make(map[string]*Node),
		Parents:       make(map[string]*Node),
		Once:          sync.Once{},
		WG:            sync.WaitGroup{},
		Status:        build.Pending,
		Label:         l,
		PriorityCount: -1,
	}
}

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

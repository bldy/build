// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"container/heap"
	"sync"

	"bldy.build/build/graph"
)

type p struct {
	q *PriorityQueue
	c *sync.Cond
}

func newP() *p {
	q := PriorityQueue([]*graph.Node{})
	return &p{
		q: &q,
		c: sync.NewCond(&sync.Mutex{}),
	}
}
func (p *p) len() int {
	return p.q.Len()
}
func (p *p) push(n *graph.Node) {
	p.c.L.Lock()

	heap.Push(p.q, n)
	p.c.Signal()
	p.c.L.Unlock()

}
func (p *p) pop() *graph.Node {
	p.c.L.Lock()
	if p.q.Len() == 0 {
		p.c.Wait()
	}
	x := heap.Pop(p.q)
	p.c.L.Unlock()
	return x.(*graph.Node)
}

type PriorityQueue []*graph.Node

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]

}

func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*graph.Node))

}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

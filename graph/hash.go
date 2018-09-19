// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"fmt"
	"sort"
	"strings"

	"bldy.build/build/racy"
)

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
	return strings.Compare(a[i].Target.Name(), a[j].Target.Name()) > 0
}

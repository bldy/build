// Package racy deals with file cryptography
package racy

import (
	"crypto/sha512"
	"hash"
	"io"
	"sync"
)

// NewHash returns a new hash.Hash
var NewHash = sha512.New

type cachedMap struct {
	mu *sync.Mutex
	m  map[string][]byte
}

// Racy is used in hashing targets.
// https://www.kernel.org/pub/software/scm/git/docs/technical/racy-git.txt
type Racy struct {
	hash.Hash

	mu *sync.Mutex
	m  map[string][]byte
}

// New takes nothing and returns a new Racy
func New() *Racy {
	return &Racy{
		NewHash(),

		&sync.Mutex{},
		make(map[string][]byte),
	}
}

func hashString(s string) []byte {
	h := NewHash()
	io.WriteString(h, s)
	return h.Sum(nil)
}

// HashString takes a hash and one or more strings and writes the strings to the
// hash.
func HashString(h io.Writer, strs ...string) {
	for _, str := range strs {
		io.WriteString(h, str)
	}
}

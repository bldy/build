// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ast defines build data structures.
package ast // import "sevki.org/build/ast"

import "sevki.org/build/token"

type File struct {
	Path  string
	Funcs []*Func
	Vars  map[string]interface{}
}

type Statement interface {
	isStatement()
}

// Position represents a index of a byte in a given file relative to the
// start of the line.
//
// 	abcd
// 	efg
//
// Given "testfile.txt" 'f' would be adressed "testfile.txt:1:1"
type Position struct {
	Line, Index int
}

// Node defines objects on file
type Node struct {
	File       string
	Start, End Position
}

// A BasicLit node represents a literal of basic type.
type BasicLit struct {
	Kind  token.Token // token.INT, token.FLOAT or token.STRING
	Value string      // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
	Node
}

// Variable type points to a variable in, or a loaded document.
type Variable struct {
	Key string
	Node
}

// Func represents a function in the ast mostly in the form of
//
// 	glob("", exclude=[], exclude_directories=1)
//
// a function can have named and anonymouse variables at the same time.
type Func struct {
	Name       string
	Params     map[string]interface{}
	AnonParams []interface{}
	Parent     *Func `json:"-"`
	Node
}

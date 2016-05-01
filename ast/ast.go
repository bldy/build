// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ast defines build data structures.
package ast // import "sevki.org/build/ast"

import (
	"errors"
	"strconv"

	"sevki.org/build/token"
)

var (
	InterfaceConversionError = errors.New("Interface conversion error")
)

type File struct {
	Path  string
	Funcs []*Func
	Vars  map[string]interface{}
}
type Decl interface {
	isDecl()
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

// Sets the start position of the node with a token
func (n *Node) SetStart(t token.Token) {
	n.Start = Position{
		Line:  t.Line,
		Index: t.Start,
	}
}

// Sets the end position of the node with a token
func (n *Node) SetEnd(t token.Token) {
	n.End = Position{
		Line:  t.Line,
		Index: t.End,
	}
}

// Variable type points to a variable in, or a loaded document.
type Variable struct {
	Key string
	Node
}

type Slice struct {
	Slice []interface{}
	Node
}

// Assignment
type Assignment struct {
	Key	string
	Value interface{}
	Node
}
func (a *Assignment)isDecl() {}

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
func (f *Func)isDecl() {}

type Map struct {
	Map map[string]interface{}
	Node
}

// A BasicLit node represents a literal of basic type.
type BasicLit struct {
	Kind  token.Type // token.INT, token.FLOAT or token.STRING
	Value string     // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
	Node
}

func NewBasicLit(t token.Token) *BasicLit {
	lit := BasicLit{
		Kind:  t.Type,
		Value: t.String(),
	}
	lit.SetStart(t)
	lit.SetEnd(t)
	return &lit
}

func (b BasicLit) Interface() interface{} {
	switch b.Kind {
	case token.Int:
		if s, err := strconv.Atoi(b.Value); err == nil {
			return s
		} else {
			return InterfaceConversionError
		}
	case token.Quote:
		return b.Value
	case token.True:
		return true
	case token.False:
		return false
	default:
		return InterfaceConversionError
	}
}

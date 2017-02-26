// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type Type

package token

type Token struct {
	Type  Type
	Text  []byte
	Line  int
	Start int
	End   int
}

type Type int

const (
	EOF Type = iota
	Error
	Newline
	String
	Space
	Int
	Float
	Hex
	LeftCurly
	RightCurly
	LeftParen
	RightParen
	LeftBrac
	RightBrac
	Quote
	Equal
	Colon
	Comma
	Semicolon
	Period
	Comment
	Plus
	Pipe
	Elipsis
	True
	False
	MultiLineString
	TargetDecl
	Func
	For
	In
	Name
	Fmt
)

func (t Token) String() string {
	return string(t.Text)
}

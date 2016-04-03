// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type Type

package token // import "sevki.org/build/token"

import (
	"errors"
	"strconv"
)

type Token struct {
	Type  Type
	Text  []byte
	Line  int
	Start int
	End   int
}

type Type int

var (
	InterfaceConversionError = errors.New("Interface conversion error")
)

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
)

func (t Token) String() string {
	return string(t.Text)
}

func (t Token) Interface() interface{} {
	switch t.Type {
	case Int:
		if s, err := strconv.Atoi(t.String()); err == nil {
			return s
		} else {
			return InterfaceConversionError
		}
	default:
		return InterfaceConversionError
	}
}

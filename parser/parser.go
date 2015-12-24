// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser // import "sevki.org/build/parser"

import (
	"fmt"
	"io"

	"sevki.org/build/token"

	"sevki.org/build/ast"
	"sevki.org/build/lexer"
)

type Parser struct {
	name     string
	path     string
	lexer    *lexer.Lexer
	state    stateFn
	peekTok  token.Token
	curTok   token.Token
	line     int
	Error    error
	Document *ast.File
	ptr      *ast.Func
	payload  map[string]interface{}
	typeName string
}

type stateFn func(*Parser) stateFn

func (p *Parser) peek() token.Token {
	return p.peekTok
}
func (p *Parser) next() token.Token {
	tok := p.peekTok
	p.peekTok = <-p.lexer.Tokens
	p.curTok = tok
	//	cal, _, _ := caller()
	//	fmt.Printf("%s\t:: %s:%s -> %s:%s\n%s\n", cal, p.curTok, p.curTok.Type, p.peekTok, p.peekTok.Type, p.path+p.name)
	if tok.Type == token.Error {
		p.errorf("%q", tok)
	}

	return tok
}

func (p *Parser) errorf(format string, args ...interface{}) {
	p.curTok = token.Token{Type: token.Error}
	p.peekTok = token.Token{Type: token.EOF}
	p.Error = fmt.Errorf(format, args...)
}

func New(name, path string, r io.Reader) *Parser {
	p := &Parser{
		name:  name,
		path:  path,
		line:  0,
		lexer: lexer.New(name, r),
		Document: &ast.File{
			Path: path,
		},
	}

	return p
}

func (p *Parser) run() {
	p.next()
	for p.state = parseBuild; p.state != nil; {
		p.state = p.state(p)
	}
}

func parseBuild(p *Parser) stateFn {
	for p.peek().Type != token.EOF {
		return parseDecl
	}
	return nil
}

func parseDecl(p *Parser) stateFn {
	switch p.peek().Type {
	case token.Func:
		return parseFunc
	case token.String:
		return parseVar
	}
	return nil
}
func parseVar(p *Parser) stateFn {
	t := p.next()
	if !p.isExpected(t, token.String) {
		return nil
	}
	if !p.isExpected(p.next(), token.Equal) {
		return nil
	}

	if p.Document.Vars == nil {
		p.Document.Vars = make(map[string]interface{})
	}

	switch p.peek().Type {
	case token.LeftBrac:
		p.Document.Vars[t.String()] = p.parseSlice()
		return parseDecl
	case token.String:
		p.Document.Vars[t.String()] = p.next()
		return parseDecl
	}

	return nil
}
func parseFunc(p *Parser) stateFn {
	t := p.next()
	if !p.isExpected(t, token.Func) {
		return nil
	}
	f := &ast.Func{
		Name: t.String(),
	}

	if p.ptr == nil {
		p.Document.Funcs = append(p.Document.Funcs, f)
		p.ptr = f

	}

	t = p.next()
	if !p.isExpected(t, token.LeftParen) {
		return nil
	}

	return parseParams
}

func parseFuncEnd(p *Parser) stateFn {
	f := p.ptr
	p.ptr = f.Parent

	p.next()

	if f.Parent == nil {
		return parseDecl
	} else {
		return parseParams
	}
}
func parseParams(p *Parser) stateFn {

	switch p.peek().Type {
	case token.Quote:
		p.ptr.AnonParams = append(p.ptr.AnonParams, p.next().String())
	case token.LeftBrac:
		p.ptr.AnonParams = append(p.ptr.AnonParams, p.parseSlice())
	case token.String:
		name := p.next().String()

		if p.ptr.Params == nil {
			p.ptr.Params = make(map[string]interface{})
		}
		if !p.isExpected(p.next(), token.Equal) {
			return nil
		}
		// named param magicication
		switch p.peek().Type {
		case token.Quote:
			p.ptr.Params[name] = p.next().String()
		case token.LeftBrac:
			p.ptr.Params[name] = p.parseSlice()
		case token.Func:
			// make a func
			f := &ast.Func{
				Name: p.next().String(),
			}
			// that func is a named param
			p.ptr.Params[name] = f
			// link params for stackjumping
			f.Parent = p.ptr
			p.ptr = f

			// parse the funkies
			t := p.next()
			if !p.isExpected(t, token.LeftParen) {
				return nil
			}
		case token.String:
			p.ptr.Params[name] = ast.Variable{Value: p.next().String()}
		default:
			return nil
		}
		return parseParams
	case token.Func:
		return parseFunc
	case token.RightParen:
		return parseFuncEnd
	default:
		return nil
	}
	return parseParams
}

func (p *Parser) parseSlice() []interface{} {
	//	fmt.Println(caller())
	if !p.isExpected(p.next(), token.LeftBrac) {
		return nil
	}
	var slc []interface{}

	for p.peek().Type != token.RightBrac {
		slc = append(slc, p.next().String())
	}
	// advance ]
	p.next()
	return slc
}

// Decode decodes a bazel/buck ast.
func (p *Parser) Decode(i interface{}) (err error) {
	p.Document = (i.(*ast.File))
	p.Document.Path = p.path
	p.run()
	if p.curTok.Type == token.Error {
		return p.Error
	}
	return nil
}

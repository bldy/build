// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser // import "sevki.org/build/parser"

import (
	"errors"
	"fmt"
	"io"

	"sevki.org/build/token"

	"sevki.org/build/ast"
	"sevki.org/build/lexer"
)

var (
	ErrConsumption = errors.New("consumption error")
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
	stack    []stateFn
}

type stateFn func(*Parser) stateFn

func (p *Parser) peek() token.Token {
	return p.peekTok
}
func (p *Parser) next() token.Token {
	tok := p.peekTok
	p.peekTok = <-p.lexer.Tokens
	p.curTok = tok

	if tok.Type == token.Error {
		p.errorf("%q", tok)
	}

	return tok
}

func (p *Parser) errorf(format string, args ...interface{}) error {
	p.curTok = token.Token{Type: token.Error}
	p.peekTok = token.Token{Type: token.EOF}
	p.Error = fmt.Errorf(format, args...)
	return p.Error
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
		if f, err :=  p.consumeFunc(); err != nil {
			p.Error = err 
			return nil 
		} else {
			p.Document.Funcs = append(p.Document.Funcs, f)
		}
		return parseDecl
	case token.String:
		return parseVar
	}
	return nil
}
func parseVar(p *Parser) stateFn {
	t := p.next()

	if err := p.expects(t, token.String); err != nil {
		p.Error = err
		return nil
	}
	if err := p.expects(p.next(), token.Equal); err != nil {
		p.Error = err
		return nil
	}

	if p.Document.Vars == nil {
		p.Document.Vars = make(map[string]interface{})
	}

	switch p.peek().Type {
	case token.LeftBrac, token.LeftCurly, token.String, token.Quote, token.True, token.False, token.Func:
		if n, err := p.consumeNode(); err != nil {
			return nil
		} else {
			p.Document.Vars[t.String()] = n
		}
	}
	if p.peek().Type == token.Plus {

		f := &ast.Func{
			Name: "addition",
		}
		f.File = p.name
		f.Line = t.Line
		f.Position = t.Start

		f.AnonParams = []interface{}{p.Document.Vars[t.String()]}

		p.Document.Vars[t.String()] = f

		for p.peek().Type == token.Plus {
			p.next()
			switch p.peek().Type {
			case token.String:
				f.AnonParams = append(
					f.AnonParams,
					ast.Variable{Key: p.next().String()},
				)
			case token.Quote:
				f.AnonParams = append(
					f.AnonParams,
					p.next().String(),
				)
			}

		}

	}

	return parseDecl
}
func (p *Parser) consumeNode() (interface{}, error) {
	switch p.peek().Type {
	case token.Quote:
		return p.next().String(), nil
	case token.LeftBrac:
		return p.consumeSlice()
	case token.LeftCurly:
		return  p.consumeMap() 
	case token.Func:
		return p.consumeFunc()
	case token.True:
		return true, nil
	case token.False:
		return false, nil
	case token.String:
		return ast.Variable{Key: p.next().String()}, nil
	default:
		return nil, ErrConsumption
	}
}
func (p *Parser) consumeParams(f *ast.Func) error {
	for {
		switch p.peek().Type {
		case token.Quote, token.LeftBrac, token.Func:
			if n, err := p.consumeNode(); err != nil {
				return err
			} else {
				f.AnonParams = append(f.AnonParams, n)
			}
		case token.String:
			t := p.next()
			if f.Params == nil {
				f.Params = make(map[string]interface{})
			}

			if err := p.expects(p.peek(), token.Colon, token.Equal); err == nil {
				switch p.next().Type {
				case token.Colon:
				case token.Equal:
					if n, err := p.consumeNode(); err != nil {
						return err
					} else {
						f.Params[t.String()] = n
					}
				}
			} else {
				return err
			}
		default:
			return ErrConsumption
		}

		if err := p.expects(p.peek(), token.RightParen, token.Comma); err == nil {
		DANGLING_COMMA:
			switch p.peek().Type {
			case token.RightParen:
				p.next()
				return nil
			case token.Comma:
				p.next()
				if p.peek().Type == token.RightParen {
					goto DANGLING_COMMA
				}
				continue
			}
		} else {
			return err
		}

	}
}
func (p *Parser) consumeMap() (map[string]interface{}, error) {

	if err := p.expects(p.next(), token.LeftCurly); err != nil {
		return nil, err
	}
	_map := make(map[string]interface{})
	for p.peek().Type != token.RightCurly {
		t := p.next()
		if err := p.expects(t, token.Quote); err != nil {
			return nil, err
		}
		if err := p.expects(p.next(), token.Colon); err != nil {
			return nil, err
		}

		if n, err := p.consumeNode(); err != nil {
			return nil, err
		} else {
			_map[t.String()] = n
		}
		if p.peek().Type == token.Comma {
			p.next()
		} else if err := p.expects(p.peek(), token.RightCurly); err != nil {
			return nil, err
		}
	}

	// advance }
	p.next()

	return _map, nil
}
func (p *Parser) consumeFunc() (*ast.Func, error) {
	t := p.next()
	if err := p.expects(t, token.Func); err != nil {
		return nil, err
	}
	f := ast.Func{
		Name: t.String(),
	}

	f.File = p.name
	f.Line = t.Line
	f.Position = t.Start

	t = p.next()
	if err := p.expects(t, token.LeftParen); err != nil {
		return nil, err
	}
	p.consumeParams(&f)
	return &f, nil
}

func (p *Parser) consumeSlice() ([]interface{}, error) {
	if err := p.expects(p.next(), token.LeftBrac); err != nil {
		return nil, err
	}
	var slc []interface{}

	for p.peek().Type != token.RightBrac {
		slc = append(slc, p.next().String())
		if p.peek().Type == token.Comma {
			p.next()
		} else if err := p.expects(p.peek(), token.RightBrac); err != nil {
			return nil, err
		}
	}

	// advance ]
	p.next()

	return slc, nil
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
 
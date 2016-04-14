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
	Decls    chan ast.Decl
	state    stateFn
	peekTok  token.Token
	curTok   token.Token
	Error    error
	Document *ast.File
}

func (p *Parser) peek() token.Token {
	return p.peekTok
}

func (p *Parser) next() token.Token {
IGNORETOKEN:
	t := <-p.lexer.Tokens

	switch t.Type {
	case token.Error:
		p.errorf("%q", t)
	case token.Newline:
		goto IGNORETOKEN
	}

	tok := p.peekTok
	p.peekTok = t
	p.curTok = tok

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
		lexer: lexer.New(name, r),
		Document: &ast.File{
			Path: path,
		},
		Decls: make(chan ast.Decl),
	}
	return p
}

func (p *Parser) run() {
	p.next()
	for p.state = parseDecl; p.state != nil; {
		p.state = p.state(p)
	}
}

type stateFn func(*Parser) stateFn

func parseDecl(p *Parser) stateFn {
	switch p.peek().Type {
	case token.Func:
		return parseFunc
	case token.String:
		return parseVar
	}
	return nil
}

func parseFunc(p *Parser) stateFn {
	if f, err := p.consumeFunc(); err != nil {
		p.Error = err
		return nil
	} else {
		p.Decls <- f
		p.Document.Funcs = append(p.Document.Funcs, f)
	}
	return parseDecl
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
	a := ast.Assignment{
		Key: t.String(),
		Value: n,
	}
	a.SetEnd(t)
	a.SetStart(p.curTok)
	p.Decls <- &a
		}
	}

	return parseDecl
}

func (p *Parser) consumeNode() (interface{}, error) {
	var r interface{}
	var err error

	switch p.peek().Type {
	case token.Quote:
		r, err = ast.NewBasicLit(p.next()), nil
	case token.True:
		r, err = ast.NewBasicLit(p.next()), nil
	case token.False:
		r, err = ast.NewBasicLit(p.next()), nil
	case token.String:
		r, err = ast.Variable{Key: p.next().String()}, nil
	case token.LeftBrac:
		r, err = p.consumeSlice()
	case token.LeftCurly:
		r, err = p.consumeMap()
	case token.Func:
		r, err = p.consumeFunc()
	default:
		return nil, ErrConsumption
	}
REPROCESS:
	switch p.peek().Type {
	case token.Plus:
		r, err = p.consumeAddFunc(r)
		goto REPROCESS
	case token.LeftBrac:
		r, err = p.consumeSliceFunc(r)
		goto REPROCESS
	}
	return r, err
}

func (p *Parser) consumeAddFunc(v interface{}) (*ast.Func, error) {
	f := &ast.Func{
		Name: "addition",
	}

	f.File = p.name
	f.SetStart(p.curTok)

	f.AnonParams = []interface{}{v}

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

	f.SetEnd(p.curTok)
	return f, nil
}

func (p *Parser) consumeSliceFunc(v interface{}) (*ast.Func, error) {

	f := &ast.Func{
		Params: make(map[string]interface{}),
	}
	f.File = p.name
	f.SetStart(p.curTok)

	// advance [
	p.next()

	f.Params["var"] = v
	if p.peek().Type == token.Colon {
		// advance :
		p.next()
		f.Name = "slice"
		f.Params["start"] = 0
		node, err := p.consumeNode()
		if err != nil {
			return nil, err
		}
		f.Params["end"] = node
		goto END
	} else if p.peek().Type == token.Int {
		node, err := p.consumeNode()
		if err != nil {
			return nil, err
		}

		if p.peek().Type == token.RightBrac {
			f.Name = "index"
			f.Params["index"] = node
			goto END
		} else if p.peek().Type == token.Colon {
			// advance :
			p.next()
		} else {
			return nil, fmt.Errorf("this is a malformed slice")
		}
		f.Name = "slice"

		node, err = p.consumeNode()
		if err != nil {
			return nil, err
		}

		f.Params["start"] = node
		if p.peek().Type == token.Int {
			node, err = p.consumeNode()
			if err != nil {
				return nil, err
			}
			f.Params["end"] = node
		} else if p.peek().Type == token.RightBrac {
			goto END
		} else {
			return nil, fmt.Errorf("this is a malformed slice")
		}

	}
END:
	// advance ]
	f.SetEnd(p.next())
	return f, nil
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
func (p *Parser) consumeMap() (*ast.Map, error) {
	t := p.next()
	_map := ast.Map{
		Value: make(map[string]interface{}),
	}
	_map.SetStart(t)

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
			_map.Value[t.String()] = n
		}
		if p.peek().Type == token.Comma {
			p.next()
		} else if err := p.expects(p.peek(), token.RightCurly); err != nil {
			return nil, err
		}
	}

	// advance }

	_map.SetEnd(p.next())
	return &_map, nil
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
	f.Start = ast.Position{
		Line:  t.Line,
		Index: t.Start,
	}

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

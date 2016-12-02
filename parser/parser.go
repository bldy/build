// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"errors"
	"fmt"
	"io"

	"bldy.build/build/ast"
	"bldy.build/build/lexer"
	"bldy.build/build/token"
)

var (
	ErrConsumption = errors.New("consumption error")
	ErrNotSlice    = errors.New("isFunc")
)

type Parser struct {
	name    string
	Path    string
	lexer   *lexer.Lexer
	Decls   chan ast.Decl
	state   stateFn
	peekTok token.Token
	curTok  token.Token
	Error   error
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
		Path:  path,
		lexer: lexer.New(name, r),
		Decls: make(chan ast.Decl),
	}
	return p
}
func (p *Parser) emit(d ast.Decl) {
	if p.Error != nil {
		p.Decls <- &ast.Error{Error: p.Error}
	} else {
		p.Decls <- d
	}
}
func (p *Parser) Run() {
	p.next()
	for p.state = parseDecl; p.state != nil; {
		p.state = p.state(p)
	}
	p.emit(nil)
	close(p.Decls)

}

type stateFn func(*Parser) stateFn

func parseDecl(p *Parser) stateFn {

	switch p.peek().Type {
	case token.Func:
		return parseFunc
	case token.String:
		return parseVar
	case token.LeftBrac:
		return parseLoop
	case token.EOF:
		return nil
	default:
		p.expects(p.peek(), token.Func, token.String, token.LeftBrac, token.EOF)
		return nil
	}
}

func parseLoop(p *Parser) stateFn {
	l := ast.Loop{}
	l.SetStart(p.next())

	if f, err := p.consumeFunc(); err != nil {
		p.Error = err
		return nil
	} else {
		l.Func = f
	}
	// advance for
	if err := p.expects(p.next(), token.For); err != nil {
		p.Error = err
		return nil
	}

	t := p.next()
	if err := p.expects(t, token.String); err != nil {
		p.Error = err
		return nil
	} else {
		l.Key = t.String()
	}

	// advance in
	if err := p.expects(p.next(), token.In); err != nil {
		p.Error = err
		return nil
	}

	if node, err := p.consumeNode(); err != nil {
		p.Error = err
		return nil
	} else {
		l.Range = node
	}

	// advance ]
	if err := p.expects(p.next(), token.RightBrac); err != nil {
		p.Error = err
		return nil
	}

	p.emit(&l)
	return parseDecl
}

func parseFunc(p *Parser) stateFn {
	if f, err := p.consumeFunc(); err != nil {
		p.Error = err
		return nil
	} else {
		p.emit(f)
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

	switch p.peek().Type {
	case token.LeftBrac, token.LeftCurly, token.String, token.Quote, token.True, token.False, token.Func:
		if n, err := p.consumeNode(); err != nil {
			return nil
		} else {
			a := ast.Assignment{
				Key:   t.String(),
				Value: n,
			}
			a.SetEnd(t)
			a.SetStart(p.curTok)
			p.emit(&a)
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
		t := p.next()
		v := ast.Variable{Key: t.String()}
		v.SetStart(t)
		v.SetEnd(t)
		r, err = &v, nil
	case token.LeftBrac:
		r, err = p.consumeSlice()
	case token.LeftCurly:
		r, err = p.consumeMap()
	case token.Func:
		r, err = p.consumeFunc()
	case token.Int:
		r, err = ast.NewBasicLit(p.next()), nil
	default:
		return nil, fmt.Errorf("unknown type %s\n%s",
			p.peek().Type,
			p.lexer.LineBuffer())
	}
REPROCESS:
	// only process modifiers at the end of the line
	if p.curTok.Line != p.peekTok.Line {
		return r, err
	}
	switch p.peek().Type {
	case token.Plus:
		r, err = p.consumeAddFunc(r)
		goto REPROCESS
	case token.LeftBrac:
		r, err = p.consumeSliceFunc(r)
		goto REPROCESS
	}
	if err != nil {
		p.Error = err
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
			t := p.next()
			v := ast.Variable{Key: t.String()}
			v.SetStart(t)
			v.SetEnd(t)

			f.AnonParams = append(
				f.AnonParams,
				&v,
			)
		case token.Quote:
			f.AnonParams = append(
				f.AnonParams,
				p.next().String(),
			)
		case token.LeftBrac:
			slc, err := p.consumeSlice()
			if err != nil {
				return nil, err
			}
			f.AnonParams = append(
				f.AnonParams,
				slc,
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

		f.Name = "slice"
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

	} else if p.peek().Type == token.Func {

		return nil, nil
	} else {
		return nil, p.errorf("unknown type consuming a slice")
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
		Map: make(map[string]interface{}),
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
			_map.Map[t.String()] = n
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

func (p *Parser) consumeSlice() (*ast.Slice, error) {
	var _slice ast.Slice

	if err := p.expects(p.peek(), token.LeftBrac); err != nil {
		return nil, err
	} else {
		_slice.SetStart(p.next())
	}

	for p.peek().Type != token.RightBrac {
		node, err := p.consumeNode()
		if err != nil {
			return nil, err
		}
		_slice.Slice = append(_slice.Slice, node)
		if p.peek().Type == token.Comma {
			p.next()
		} else if err := p.expects(p.peek(), token.RightBrac); err != nil {
			return nil, err
		}
	}

	// advance ]
	_slice.SetEnd(p.next())

	return &_slice, nil
}

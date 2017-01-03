// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"runtime"
	"strings"

	"bldy.build/build/token"
)

func caller() (call string, file string, line int) {
	var caller uintptr
	caller, file, line, _ = runtime.Caller(2)
	name := strings.Split(runtime.FuncForPC(caller).Name(), ".")
	callName := name[len(name)-1]
	return strings.Trim(callName, "parse"), file, line

}

func arrow(buf string, tok token.Token) string {
	ret := ""
	for i := 0; i < len(string(buf)); i++ {
		if i >= tok.Start && i <= tok.End {
			ret += "^"
			continue
		} else {
			ret += " "

		}
		switch i {
		case tok.Start - 1, tok.Start - 2, tok.Start - 3:
			ret += ">"
			break
		case tok.End + 1, tok.End + 2, tok.End + 3:
			ret += "<"
			break
		default:
			ret += " "
		}
	}
	return ret
}

func (p *Parser) expects(tok token.Token, expected ...token.Type) error {
	for _, t := range expected {
		if t == tok.Type {
			return nil
		}
	}
	name, _, _ := caller()
	errf := "%s:%d: While parsing %s were expecting %s but got %s."
	errf += "\n%s\n%s"
	return p.errorf(errf,
		p.Path,
		tok.Line,
		name,
		expected,
		p.curTok.Type,
		strings.Trim(p.lexer.LineBuffer(), "\n"),
		arrow(p.lexer.LineBuffer(), tok),
	)
}


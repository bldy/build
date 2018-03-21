// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package label // import "bldy.build/build/label"
import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"unicode"
)

const (
	EOF = rune(26)
)

// Label represents a perforce label
// we plan on adding more providers
type Label struct {
	Package *string
	Name    string
}

func (lbl Label) String() string {
	if lbl.Package != nil {
		return fmt.Sprintf("//%s:%s", *lbl.Package, lbl.Name)
	}
	return fmt.Sprintf(":%s", lbl.Name)
}

func Package(s string) *string {
	x := s
	return &x
}

type parser struct {
	err          error
	s            string
	q            chan byte
	i            int
	lastNonPrint int
	firstLetter  int
	packageName  []rune
	targetName   []rune
	state        stateFn
	file         bool
}

func newParser(a string) *parser {
	p := parser{
		s:            a,
		q:            make(chan byte),
		state:        nil,
		lastNonPrint: 0,
		file:         false,
	}
	return &p
}
func (p *parser) current() rune {
	return rune(p.s[p.i])
}
func (p *parser) peek() rune {
	return rune(p.s[p.i+1])
}

func (p *parser) next() rune {

	p.i++
	if i := p.i; i >= len(p.s) {
		return EOF
	}
	x := rune(p.s[p.i])

	switch x {
	case '/', ':':
		p.lastNonPrint = p.i
	}
	return x
}

func (p *parser) backoff() {
	//	d := p.i - p.firstLetter
	p.i = p.lastNonPrint
	//log.Println(string(p.packageName[:]), len(p.packageName)-d)

	//p.packageName = p.packageName[:len(p.packageName)-d]
}

func (p *parser) Error(link, format string, args ...interface{}) {
	buf := bytes.NewBuffer(nil)
	msg := make([]byte, len(p.s))
	for i := 0; i < len(p.s); i++ {
		if i < p.i {
			msg[i] = '>'
		} else if i > p.i {
			msg[i] = '<'
		} else {
			msg[i] = '^'
		}
	}
	fmt.Fprintf(buf, format, args...)

	fmt.Fprintln(buf)
	fmt.Fprintln(buf, p.s)
	fmt.Fprintln(buf, string(msg))
	fmt.Fprintf(buf, "char %d\n", p.i)

	if link != "" {
		fmt.Fprintf(buf, "please refer to %s for more details.", link)
	}

	p.err = errors.New(buf.String())
}

type stateFn func(*parser) stateFn

func parseLabel(p *parser) stateFn {
	switch p.current() {
	case '/':
		return parseAbsolute
	case ':':
		return parsePackageName
	default:
		p.i--
		return parsePackageName
	}
}

func parseAbsolute(p *parser) stateFn {
	for i := 0; i < 1; i++ {
		if c := p.next(); c != '/' {
			p.Error("", "was expecting '/' got '%c' instead", c)
			return nil
		}
	}
	return parsePackageName
}

func parseTargetName(p *parser) stateFn {
	for {
		c := p.next()
		if c == EOF {
			return nil
		}
		if isValidTargetNameChar(c) {
			p.targetName = append(p.targetName, c)
		} else {
			p.Error("https://docs.bazel.build/versions/master/build-ref.html#name", "target names can't have '%c' characters", c)
			return nil
		}
	}
}

func parsePackageName(p *parser) stateFn {
	first := true
	for {
		c := p.next()

		if first && c == '_' {
			p.Error("https://docs.bazel.build/versions/master/build-ref.html#package-names-package-name", "package names can't start with underscore")
			return nil
		}
		inValid := !isValidPackageChar(c)

		switch c {
		case '.':
			p.file = true
			p.backoff()
			return parseTargetName
		case ':':
			return parseTargetName
		case EOF:
			return nil
		}

		if inValid {
			p.Error("https://docs.bazel.build/versions/master/build-ref.html#package-names-package-name", "package names can't have '%c'", c)
			return nil
		}
		p.packageName = append(p.packageName, c)
		first = false
	}
}

//Package names must be composed entirely of characters drawn from the set A-Z, a–z, 0–9, '/', '-', '.', and '_', and cannot start with a slash.
//
// https://docs.bazel.build/versions/master/build-ref.html#package-names-package-name
func isValidPackageChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '/' || r == '-' || r == '.' || r == '_'
}

//Target names must be composed entirely of characters drawn from the set a–z, A–Z, 0–9, and the punctuation symbols _/.+-=,@~.
//
// https://docs.bazel.build/versions/master/build-ref.html#name
func isValidTargetNameChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '/' || r == '.' || r == '+' || r == '-' || r == '=' || r == ',' || r == '@' || r == '~'
}

// Parse takes a string and returns a bazel label
func Parse(s string) (*Label, error) {
	p := newParser(s)
	for p.state = parseLabel; p.state != nil; {
		p.state = p.state(p)
	}
	if p.err != nil {
		return nil, p.err
	}
	l := new(Label)
	pkgName := string(p.packageName)
	l.Package = Package(pkgName)

	l.Name = string(p.targetName)
	if l.Name == "" {
		_, l.Name = path.Split(pkgName)
	}
	return l, nil
}

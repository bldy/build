// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package label // import "bldy.build/build/label"
import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/google/skylark"
)

var (
	ErrInvalidLabel = errors.New("label couldn't be parsed")
)

// Label represents a perforce label
// we plan on adding more providers
type Label string

func New(pkg, name string) Label {
	return Label(fmt.Sprintf("//%s:%s", pkg, name))
}

func (l Label) Repo() string {
	if f, err := l.Split(); err == nil {
		if r, ok := f["repo"]; ok {
			return r
		}
	}
	return ""
}

func (l Label) Package() string {
	if f, err := l.Split(); err == nil {
		if r, ok := f["package"]; ok {
			return r
		}
	}
	return ""
}

func (l Label) Name() string {
	if f, err := l.Split(); err == nil {
		if r, ok := f["target"]; ok {
			return r
		}
	}
	return ""
}

func (l Label) IsAbs() bool {
	return strings.HasPrefix(string(l), "//")
}

func Package(s string) *string {
	x := s
	return &x
}

func (l Label) Valid() error {
	_, err := l.Split()
	return err
}

func Parse(s string) (Label, error) {
	l := Label(s)
	_, err := l.Split()
	return l, err
}

// Split splits an label.Label into package and label pair
//
// 	<label> := //<package name>:<target name>
//
// 	<package name> :=
//
//	https://docs.bazel.build/versions/master/build-ref.html#package-names-package-name
//
//	The name of a package is the name of the directory containing its BUILD file,
//	relative to the top-level directory of the source tree. For example: my/app.
//	Package names must be composed entirely of characters drawn from the
//	set A-Z, a–z, 0–9, '/', '-', '.', and '_', and cannot start with a slash.
//
//	For a language with a directory structure that is significant to its module system
//	(e.g. Java), it is important to choose directory names that are valid identifiers in the language.
//
//	Although Bazel allows a package at the build root (e.g. //:foo), this is not advised
//	and projects should attempt to use more descriptively named packages.
//
//	Package names may not contain the substring //, nor end with a slash.
//
// 	<target name> :=
//
// 	https://docs.bazel.build/versions/master/build-ref.html#name
//
//	Target names must be composed entirely of characters drawn from
//	the set a–z, A–Z, 0–9, and the punctuation symbols _/.+-=,@~.
//	Do not use .. to refer to files in other packages; use //packagename:filename instead.
//	Filenames must be relative pathnames in normal form, which
//	means they must neither start nor end with a slash (e.g. /foo and foo/ are forbidden)
//	nor contain multiple consecutive slashes as path separators (e.g. foo//bar).
//	Similarly, up-level references (..) and current-directory references (./) are forbidden.
//	The sole exception to this rule is that a target name may consist of exactly '.'.
//
//	While it is common to use / in the name of a file target, we recommend
//	that you avoid the use of / in the names of rules. Especially when the shorthand form
//	of a label is used, it may confuse the reader. The label //foo/bar/wiz is always a
//	shorthand for //foo/bar/wiz:wiz, even if there is no such package foo/bar/wiz;
//	it never refers to //foo:bar/wiz, even if that target exists.
//
//	However, there are some situations where use of a slash is convenient, or sometimes
//	even necessary. For example, the name of certain rules must match their principal
//	source file, which may reside in a subdirectory of the package.
func (l Label) Split() (map[string]string, error) {
	var (
		fullLabel   = regexp.MustCompile("(@(?P<repo>[[:alnum:]-_.]*[[:alnum:]-_./]*))?//(?P<package>[[:alnum:]-_.]*[[:alnum:]-_./]*):(?P<target>[[:alnum:]_/.+-=,@~.]*)+")
		packageOnly = regexp.MustCompile("//(?P<package>[[:alnum:]-_.][[:alnum:]-_./]+)")
		targetOnly  = regexp.MustCompile(":?(?P<target>[[:alnum:]_/?.+-=,@~.]*)+")

		//filename
		fileName = regexp.MustCompile(`((?P<package>\A[[:alnum:]]+[[:alnum:]/]+)/)*(?P<target>[[:alnum:]]+.[[:alnum:]]+)`)
		file     = regexp.MustCompile(`(?P<target>[[:alnum:]]+.[[:alnum:]]+)`)
	)
	s := string(l)
	_ = targetOnly.MatchString(s)
	matches := [][]string{}
	names := []string{}

	switch {
	case fullLabel.MatchString(s):
		matches = fullLabel.FindAllStringSubmatch(s, 1)
		names = fullLabel.SubexpNames()
		break
	case packageOnly.MatchString(s):
		matches = packageOnly.FindAllStringSubmatch(s, 1)
		names = packageOnly.SubexpNames()
		break
	case targetOnly.MatchString(s):
		matches = targetOnly.FindAllStringSubmatch(s, 1)
		names = targetOnly.SubexpNames()
		break
	case fileName.MatchString(s):
		matches = fileName.FindAllStringSubmatch(s, 1)
		names = fileName.SubexpNames()
		break
	case file.MatchString(s):
		matches = fileName.FindAllStringSubmatch(s, 1)
		names = fileName.SubexpNames()
	default:
		return nil, ErrInvalidLabel
	}
	if len(matches) < 1 {
		return nil, ErrInvalidLabel
	}
	parts := matches[0]
	frags := make(map[string]string)
	for i, match := range names {
		frags[match] = parts[i]
	}
	if _, ok := frags["target"]; !ok {
		if pkg, ok := frags["package"]; ok {
			_, frags["target"] = path.Split(pkg)
		} else {
			return nil, fmt.Errorf("label in incorrect format %q", s)
		}
	}

	return frags, nil
}

func (l Label) Type() string        { return "label" }
func (l Label) Freeze()             {}
func (l Label) Truth() skylark.Bool { return skylark.Bool(true) }
func (l Label) Hash() (uint32, error) {
	s := l.String()
	var h uint32
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h, nil
}
func (l Label) String() string {

	return string(l)
}

func (l Label) Attr(attr string) (skylark.Value, error) {
	if attr == "name" {
		attr = "target"
	}
	if frags, err := l.Split(); err == nil {
		if a, ok := frags[attr]; ok {
			return skylark.String(a), nil
		}
	}

	return nil, fmt.Errorf("label(%q) has no attribute called %q", l, attr)
}

func (l Label) AttrNames() []string { panic("not implemented") }

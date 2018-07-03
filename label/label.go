// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package label // import "bldy.build/build/label"
import (
	"errors"
	"fmt"
	"path"
	"regexp"

	"github.com/google/skylark"
)

var (
	ErrInvalidLabel = errors.New("label couldn't be parsed")
)

// Label represents a perforce label
// we plan on adding more providers
type Label struct {
	Package *string
	Name    string
}

func Package(s string) *string {
	x := s
	return &x
}

// Parse splits an label.Label into package and label pair
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
func Parse(s string) (*Label, error) {
	var (
		fullLabel   = regexp.MustCompile("//(?P<package>[[:alnum:]-_.]*[[:alnum:]-_./]*):(?P<target>[[:alnum:]_/.+-=,@~.]*)+")
		packageOnly = regexp.MustCompile("//(?P<package>[[:alnum:]-_.][[:alnum:]-_./]+)")
		targetOnly  = regexp.MustCompile(":?(?P<target>[[:alnum:]_/?.+-=,@~.]*)+")

		//filename
		fileName = regexp.MustCompile(`(?P<package>\A[[:alnum:]]+[[:alnum:]/]+?)/(?P<target>[[:alnum:]]+.[[:alnum:]]+)`)
	)
	_ = targetOnly.MatchString(s)
	l := &Label{}
	matches := [][]string{}
	names := []string{}
	switch {
	case fileName.MatchString(s):
		matches = fileName.FindAllStringSubmatch(s, 1)
		names = fileName.SubexpNames()
	case fullLabel.MatchString(s):
		matches = fullLabel.FindAllStringSubmatch(s, 1)
		names = fullLabel.SubexpNames()
	case packageOnly.MatchString(s):
		matches = packageOnly.FindAllStringSubmatch(s, 1)
		names = packageOnly.SubexpNames()
	case targetOnly.MatchString(s):
		matches = targetOnly.FindAllStringSubmatch(s, 1)
		names = targetOnly.SubexpNames()
	default:
		return nil, ErrInvalidLabel
	}
	if len(matches) < 1 {
		return nil, ErrInvalidLabel
	}
	frags := matches[0]
	for i, name := range names {
		switch name {
		case "package":
			l.Package = &frags[i]
		case "target":
			l.Name = frags[i]
		}
	}

	if l.Name == "" {
		if l.Package == nil {
			return nil, fmt.Errorf("label in incorrect format %q", s)
		}
		_, l.Name = path.Split(*l.Package)
	}
	return l, nil
}

func (lbl Label) Type() string        { return "label" }
func (lbl Label) Freeze()             {}
func (lbl Label) Truth() skylark.Bool { return skylark.Bool(true) }
func (lbl Label) Hash() (uint32, error) {
	s := lbl.String()
	var h uint32
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h, nil
}
func (lbl Label) String() string {
	if lbl.Package == nil {
		return fmt.Sprintf("//%s:%s", ".", lbl.Name)
	}
	return fmt.Sprintf("//%s:%s", *lbl.Package, lbl.Name)
}

func (lbl Label) Attr(name string) (skylark.Value, error) {
	switch name {
	case "name":
		return skylark.String(lbl.Name), nil
	default:
		return nil, fmt.Errorf("label has no attribute called %q", name)
	}
}

func (lbl Label) AttrNames() []string {
	panic("not implemented")
}

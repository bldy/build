// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package label // import "bldy.build/build/label"
import (
	"fmt"
	"path"
	"strings"
)

const buildfile = "BUILD"

// Label represents a perforce label
// we plan on adding more providers
type Label struct {
	Package *string
	Name    string
}

func split(s string, c string, cutc bool) (string, string) {
	i := strings.Index(s, c)
	if i < 0 {
		return s, ""
	}
	if cutc {
		return s[:i], s[i+len(c):]
	}
	return s[:i], s[i:]
}

func (lbl Label) String() string {
	if lbl.Package != nil {
		return fmt.Sprintf("//%s:%s", *lbl.Package, lbl.Name)
	} else {
		return fmt.Sprintf(":%s", lbl.Name)

	}
}

func Package(s string) *string {
	x := s
	return &x
}

// Parse takes a string and returns a bazel label
func Parse(s string) (*Label, error) {
	lbl := new(Label)
	switch {
	case strings.HasPrefix(s, "//"):
		s = s[2:]
		pkg, name := split(s, ":", true)

		lbl.Name = name
		if pkg == "" {
			lbl.Package = Package(".") // this is the root of the project
		} else {
			lbl.Package = Package(pkg)
		}
	case strings.HasPrefix(s, ":"):
		s = s[1:]
		fallthrough
	default:
		lbl.Name = s

		lbl.Package = nil

	}

	if lbl.Name == "" && lbl.Package != nil {
		lbl.Name = path.Base(*lbl.Package)
	}

	return lbl, nil
}

// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package label

import (
	"testing"
)

// Test that all valid Labels get parsed into proper (package, target) pairs.
func TestTargetLabelParse(t *testing.T) {
	tests := []struct {
		name   string
		label  string
		repo   string
		pkg    string
		target string
	}{
		// All targets belong to exactly one package. The name of a target is called its label,
		// and a typical label in canonical form looks like this:
		{"full", "//my/app/main:app_binary", "", ("my/app/main"), "app_binary"},
		// Each label has two parts, a package name (my/app/main) and a target name (app_binary).
		// Every label uniquely identifies a target. Labels sometimes appear in other forms; when the
		// colon is omitted, the target name is assumed to be the same as the last component of the package
		// name, so these two labels are equivalent:
		{"with name", "//my/app:app", "", "my/app", "app"},
		{"omit name", "//my/app", "", "my/app", "app"},
		// Within a BUILD file, the package-name part of label may be omitted, and optionally the colon too.
		// So within the BUILD file for package my/app (i.e. //my/app:BUILD), the following "relative" labels are all equivalent:
		{"omit package", ":app", "", "", "app"},
		{"omit package and colon", "app", "", "", "app"},
		//
		{"filename", "empty/empty.bzl", "", "", "empty/empty.bzl"},
		//
		{"filename with out package", ":execute.bzl", "", "", "execute.bzl"},
		{"filename with out package", ":execute.bzl", "", "", "execute.bzl"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := Label(test.label)
			if err := l.Valid(); err != nil {
				t.Log(err)
				t.FailNow()
			}
			repo, pkg, target := l.Repo(), l.Package(), l.Name()

			if expected, got := test.target, target; expected != got {
				t.Logf("splitting %q: was expecting name %q got %q instead", test.label, expected, got)
				t.Fail()
			}
			if expected, got := test.pkg, pkg; expected != got {
				t.Logf("splitting %q: was expecting package %q got %q instead", test.label, expected, got)
				t.Fail()
			}
			if expected, got := test.repo, repo; expected != got {
				t.Logf("splitting %q: was expecting name %q got %q instead", test.label, expected, got)
				t.Fail()
			}
		})

	}
}

// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package label

import (
	"testing"

	"sevki.org/x/pretty"
)

// Test that all valid Labels get parsed into proper (package, target) pairs.
func TestTargetLabelParse(t *testing.T) {
	tests := []struct {
		name    string
		Label   string
		Package *string
		Name    string
	}{
		// All targets belong to exactly one package. The name of a target is called its label,
		// and a typical label in canonical form looks like this:
		{"full", "//my/app/main:app_binary", Package("my/app/main"), "app_binary"},
		// Each label has two parts, a package name (my/app/main) and a target name (app_binary).
		// Every label uniquely identifies a target. Labels sometimes appear in other forms; when the
		// colon is omitted, the target name is assumed to be the same as the last component of the package
		// name, so these two labels are equivalent:
		{"with name", "//my/app:app", Package("my/app"), "app"},
		{"omit name", "//my/app", Package("my/app"), "app"},
		// Within a BUILD file, the package-name part of label may be omitted, and optionally the colon too.
		// So within the BUILD file for package my/app (i.e. //my/app:BUILD), the following "relative" labels are all equivalent:
		{"omit package", ":app", nil, "app"},
		{"omit package and colon", "app", nil, "app"},
		//
		{"filename", "empty/empty.bzl", Package("empty"), "empty.bzl"},
		//
		{"filename with out package", ":execute.bzl", nil, "execute.bzl"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l, err := Parse(test.Label)
			if err != nil {
				t.Error(err)
				return
			}
			{
				test, got := test.Package, l.Package
				if test != nil {
					if got == nil {
						t.Errorf("wasn't expecting a nil package")
						t.Fail()
					}
				}
				if (test != nil && got != nil) && *test != *got {
					t.Errorf("tested package %q, got %q", *test, *got)
					t.Errorf(pretty.JSON(l))
					t.Fail()
				}
			}
			{
				if test, got := test.Name, l.Target; test != got {
					t.Errorf("expected %q, got %q", test, got)
				}
			}
		})

	}
}

// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package label

import (
	"bytes"
	"os"
	"testing"
)

// Test that all valid labels get parsed into proper (package, target) pairs.
func TestLoad(t *testing.T) {
	tbl := []struct {
		name string
		s    string
		b    []byte
		err  error
	}{
		{
			name: "full",
			s:    "//label/testdata:x",
			b:    []byte("test\n"),
		},
		{
			name: "nopackage",
			s:    ":x",
			b:    []byte("test\n"),
		},
		{
			name: "file",
			s:    "test.bldy",
			b:    []byte("bldy\n"),
		},
	}
	for _, test := range tbl {
		t.Run(test.name, func(t *testing.T) {
			wd, _ := os.Getwd()
			defer os.Chdir(wd)
			os.Chdir("testdata")
			bytz, err := Load(test.s)
			if err != test.err {
				t.Fail()
				t.Logf("was expecting %s got %s intead", test.err, err)
			}
			if bytes.Compare(test.b, bytz) != 0 {
				t.Fail()
				t.Logf("was expecting %s got %s intead", test.b, bytz)
			}
		})
	}
}

// Test that all valid Labels get parsed into proper (package, target) pairs.
func TestTargetLabelParse(t *testing.T) {
	tests := []struct {
		name    string
		Label   string
		Package string
		Name    string
	}{
		// These should all be equivalent
		{"full", "//label:label", "label", "label"},
		{"notarget", "//label:", "label", "label"},
		{"nopackage", "//label", "label", "label"},
		{"currenttarg", ":label", "label", "label"},
		{"justname", "label", "label", "label"},
		// This might not be valid if specified in a BUILD file, but the rules
		// say we should get a result
		{"empty", "", "label", "label"},
		// test a tiny target
		{"wildcard in current", ":*", "label", "*"},
		{"wildcard", "*", "label", "*"},
		// test root
		{"root target", "//:label", ".", "label"},
		{"root and dot", "//.:label", ".", "label"},
		// test name with extension
		{"name with extension", "x.bldy", "label", "x.bldy"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l, _ := Parse(test.Label)

			if test, got := test.Package, l.Package; test != got {
				t.Errorf("tested package %q, got %q", test, got)
			}
			if test, got := test.Name, l.Name; test != got {
				t.Errorf("tested target %q, got %q", test, got)
			}
		})

	}
}

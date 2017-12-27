// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package url // import "bldy.build/build/url"
import (
	"bytes"
	"os"
	"testing"
)

// Test that all valid urls get parsed into proper (package, target) pairs.
func TestLoad(t *testing.T) {
	tbl := []struct {
		name string
		s    string
		b    []byte
		err  error
	}{
		{
			name: "full",
			s:    "//url/testdata:x",
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

// Test that all valid urls get parsed into proper (package, target) pairs.
func TestTargetURLParse(t *testing.T) {
	tests := []struct {
		name    string
		URL     string
		Package string
		Target  string
	}{
		// These should all be equivalent
		{"full", "//url:url", "url", "url"},
		{"notarget", "//url:", "url", "url"},
		{"nopackage", "//url", "url", "url"},
		{"currenttarg", ":url", "url", "url"},
		{"justname", "url", "url", "url"},
		// This might not be valid if specified in a BUILD file, but the rules
		// say we should get a result
		{"empty", "", "url", "url"},
		// test a tiny target
		{"wildcard in current", ":*", "url", "*"},
		{"wildcard", "*", "url", "*"},
		// test root
		{"root target", "//:url", ".", "url"},
		{"root and dot", "//.:url", ".", "url"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url, _ := Parse(test.URL)

			if test, got := test.Package, url.Package; test != got {
				t.Errorf("testected package %q, got %q", test, got)
			}
			if test, got := test.Target, url.Target; test != got {
				t.Errorf("testected target %q, got %q", test, got)
			}
		})

	}
}

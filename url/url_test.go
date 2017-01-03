// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package url // import "bldy.build/build/url"
import "testing"

// Test that all valid urls get parsed into proper (package, target) pairs.
func TestTargetURLParse(t *testing.T) {
	tbl := []struct {
		URL     string
		Package string
		Target  string
	}{
		// These should all be equivalent
		{"//url:url", "url", "url"},
		{"//url:", "url", "url"},
		{"//url", "url", "url"},
		{":url", "url", "url"},
		{"url", "url", "url"},
		// This might not be valid if specified in a BUILD file, but the rules
		// say we should get a result
		{"", "url", "url"},
		// test a tiny target
		{":*", "url", "*"},
		{"*", "url", "*"},
		// test root
		{"//:url", ".", "url"},
		{"//.:url", ".", "url"},
	}

	for _, exp := range tbl {
		url := Parse(exp.URL)

		if exp, got := exp.Package, url.Package; exp != got {
			t.Fatalf("exp: %s, got: %s", exp, got)
		}
		if exp, got := exp.Target, url.Target; exp != got {
			t.Fatalf("exp: %s, got: %s", exp, got)
		}
	}
}

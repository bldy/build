// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser // import "sevki.org/build/parser"

import (
	"log"
	"os"
	"testing"

	"path/filepath"

	"strings"

	"sevki.org/build/ast"
	_ "sevki.org/build/targets/cc"
	"sevki.org/lib/prettyprint"
)

// func TestHarvey(t *testing.T) {

// 	var doc ast.BuildFile
// 	ks, _ := os.Open("../tests/harvey/BUILD")
// 	if err := New("BUILD", ks).Decode(&doc); err != nil {
// 		t.Error(err.Error())

// 		if err != nil {
// 			t.Error(err)
// 		}

// 	} else {
// 		log.Printf(prettyprint.AsJSON(doc))
// 	}

// }

// func TestBazel(t *testing.T) {

// 	var doc ast.BuildFile
// 	ks, _ := os.Open("../tests/c/BUILD")
// 	if err := New("BUILD", ks).Decode(&doc); err != nil {
// 		t.Error(err.Error())

// 		if err != nil {
// 			t.Error(err)
// 		}

// 	} else {
// 		log.Printf(prettyprint.AsJSON(doc))
// 	}

// }

// func TestBuck(t *testing.T) {

// 	var doc ast.BuildFile
// 	ks, _ := os.Open("../tests/c/BUCK")
// 	if err := New("BUILD", ks).Decode(&doc); err != nil {
// 		t.Error(err.Error())

// 		if err != nil {
// 			t.Error(err)
// 		}

// 	} else {
// 		log.Printf(prettyprint.AsJSON(doc))
// 	}

// }

// func TestLibXString(t *testing.T) {

// 	var doc ast.Document
// 	ks, _ := os.Open("../tests/libxstring/BUILD")
// 	ts, _ := filepath.Abs(ks.Name())
// 	dir := strings.Split(ts, "/")
// 	if err := New("BUILD", "/"+filepath.Join(dir[:len(dir)-1]...), ks).Decode(&doc); err != nil {
// 		t.Error(err.Error())
// 		if err != nil {
// 			t.Error(err)
// 		}

// 	} else {
// 		log.Printf(prettyprint.AsJSON(doc))
// 	}

// }

func TestPreProcessor(t *testing.T) {

	var doc ast.File
	ks, _ := os.Open("../tests/libxstring/BUILD")
	ts, _ := filepath.Abs(ks.Name())
	dir := strings.Split(ts, "/")
	if err := New("BUILD", "/"+filepath.Join(dir[:len(dir)-1]...), ks).Decode(&doc); err != nil {
		t.Error(err.Error())
		if err != nil {
			t.Error(err)
		}

	} else {
		var pp PreProcessor

		log.Printf(prettyprint.AsJSON(pp.Process(&doc)))
	}

}

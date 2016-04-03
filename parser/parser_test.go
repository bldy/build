// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser // import "sevki.org/build/parser"

import (
	"fmt"
	"log"
	"os"
	"testing"

	"path/filepath"

	"strings"

	"sevki.org/build/ast"
	_ "sevki.org/build/targets/cc"
)

func readAndParse(n string) (*ast.File, error) {

	var doc ast.File
	ks, err := os.Open(n)
	if err != nil {
		return nil, fmt.Errorf("opening file: %s\n", err.Error())
	}
	ts, _ := filepath.Abs(ks.Name())
	dir := strings.Split(ts, "/")
	p := New("BUILD", "/"+filepath.Join(dir[:len(dir)-1]...), ks)

	if err := p.Decode(&doc); err != nil {

		if err != nil {
			return nil, fmt.Errorf("decoding file: %s\n", err)
		}

	}

	return &doc, nil

}

func TestParseSingleVar(t *testing.T) {
	doc, err := readAndParse("tests/var.BUILD")
	if err != nil {
		t.Error(err)
	}

	if doc.Vars["UNDESIRED"].(ast.BasicLit).Interface().(string) != "-fplan9-extensions" {
		log.Fatal(doc.Vars["UNDESIRED"])
	
		t.Fail()
	}
}

func TestParseBoolVar(t *testing.T) {
	doc, err := readAndParse("tests/bool.BUILD")
	if err != nil {
		t.Error(err)
	}

	if !doc.Vars["TRUE_BOOL"].(ast.BasicLit).Interface().(bool) {
		log.Fatal(doc.Vars["TRUE_BOOL"])
		t.Fail()
	}

}

func TestParseSlice(t *testing.T) {

	strs := []string{
		"-Wall",
		"-ansi",
		"-Wno-unused-variable",
		"-pedantic",
		"-Werror",
		"-c",
	}

	doc, err := readAndParse("tests/slice.BUILD")
	if err != nil {
		t.Error(err)
	}
 
	v := doc.Vars["C_FLAGS"]
	switch v.(type) {
	case []interface{}:
		for i, x := range v.([]interface{}) {
			if strs[i] != x.(string) {
				t.Fail()
			}
		}

	default:
		t.Fail()

	}

}

func TestParseSliceWithOutComma(t *testing.T) {
	strs := []string{
		"-Wall",
		"-ansi",
		"-Wno-unused-variable",
		"-pedantic",
		"-Werror",
		"-c",
	}

	doc, err := readAndParse("tests/sliceWithOutLastComma.BUILD")
	if err != nil {
		t.Error(err)
	}

	v := doc.Vars["C_FLAGS"]
	switch v.(type) {
	case []interface{}:
		for i, x := range v.([]interface{}) {
			if strs[i] != x.(string) {
				t.Fail()
			}
		}

	default:
		t.Fail()
	}

}

func TestParseVarFunc(t *testing.T) {

	doc, err := readAndParse("tests/varFunc.BUILD")
	if err != nil {
		t.Error(err)
	}
	v := doc.Vars["XSTRING_SRCS"]
	switch v.(type) {
	case *ast.Func:

		f := v.(*ast.Func)
		if f.Name != "glob" {
			t.Fail()
		}
		q := f.AnonParams[0].([]interface{})

		if q[0] != "*.c" {
			t.Fail()
		}

	default:
		t.Fail()
	}
}

func TestParseAddition(t *testing.T) {

	doc, err := readAndParse("tests/addition.BUILD")
	if err != nil {
		t.Error(err)
	}

	v := doc.Vars["XSTRING_SRCS"]
	switch v.(type) {
	case *ast.Func:
		f := v.(*ast.Func)
		if f.Name != "addition" {
			t.Fail()
		}

		if f.AnonParams[0].(ast.Variable).Key != "CC_FLAGS" {
			t.Fail()
		}

	default:
		t.Fail()
	}

	v = doc.Vars["GOO_SRCS"]
	switch v.(type) {
	case *ast.Func:
		f := v.(*ast.Func)
		if f.Name != "addition" {
			t.Fail()
		}

		if f.AnonParams[0].(ast.Variable).Key != "CC_FLAGS" {
			t.Fail()
		}

	default:
		t.Fail()
	}
}

func TestParseMap(t *testing.T) {
	doc, err := readAndParse("tests/map.BUILD")
	if err != nil {
		t.Error(err)
		return
	}

	v, ok := doc.Vars["SOME_MAP"]
	if !ok {
		t.Fail()
		return
	}
	switch v.(type) {
	case map[string]interface{}:
		f := v.(map[string]interface{})
		if f["bla"] != "b" && f["foo"] != "p" {
			t.Fail()
		}
		return
	}
}
func TestParseMapInFunc(t *testing.T) {
	doc, err := readAndParse("tests/mapinfunc.BUILD")
	if err != nil {
		t.Error(err)
		return
	}

	if doc.Funcs[0].Params["exports"].(*ast.Map).Value["bla"].(ast.BasicLit).Interface().(string) != "b" {
			t.Fail()
	}
	if doc.Funcs[0].Params["deps"].([]interface{})[0] != ":libxstring" {
		t.Fail()
	}
	if doc.Funcs[0].Params["name"].(ast.BasicLit).Interface().(string) != "test" {
		t.Fail()
	}
	if doc.Funcs[0].Params["srcs"].([]interface{})[0] != "tests/test.c" {
		t.Fail()
	}
}
func TestParseFunc(t *testing.T) {

	doc, err := readAndParse("tests/func.BUILD")
	if err != nil {
		t.Error(err)
	}

	if doc.Funcs[0].Params["copts"].(ast.Variable).Key != "C_FLAGS" {
		t.Fail()
	}
	if doc.Funcs[0].Params["deps"].([]interface{})[0] != ":libxstring" {
		t.Fail()
	}
	if doc.Funcs[0].Params["name"].(ast.BasicLit).Interface().(string) != "test" {
		t.Fail()
	}
	if doc.Funcs[0].Params["srcs"].([]interface{})[0] != "tests/test.c" {
		t.Fail()
	}
}

func TestParseSmileyFunc(t *testing.T) {

	doc, err := readAndParse("tests/☺☹☻.BUILD")
	if err != nil {
		t.Error(err)
	}

	if doc.Funcs[0].Params["deps"].([]interface{})[0] != ":☹☻☺" {
		t.Fail()
	}
	if doc.Funcs[0].Params["name"].(ast.BasicLit).Interface().(string) != "☹☺☻" {
		t.Fail()
	}
	if doc.Funcs[0].Params["srcs"].([]interface{})[0] != "☺☹☻.c" {
		t.Fail()
	}
}

func TestParseHarvey(t *testing.T) {

	_, err := readAndParse("tests/harvey.BUILD")
	if err != nil {
		t.Error(err)
	}

}

func TestParseFull(t *testing.T) {

	_, err := readAndParse("tests/full.BUILD")
	if err != nil {
		t.Error(err)
	}

}

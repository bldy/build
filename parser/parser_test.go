// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser // import "sevki.org/build/parser"

import (
	"fmt"
	"os"
	"testing"

	"path/filepath"

	"strings"

	"sevki.org/build/ast"
	_ "sevki.org/build/targets/cc"
	"sevki.org/build/token"
	"sevki.org/lib/prettyprint"
)

func readAndParse(n string) (chan ast.Decl, error) {

	ks, err := os.Open(n)
	if err != nil {
		return nil, fmt.Errorf("opening file: %s\n", err.Error())
	}
	ts, _ := filepath.Abs(ks.Name())
	dir := strings.Split(ts, "/")
	p := New("BUILD", "/"+filepath.Join(dir[:len(dir)-1]...), ks)

	go p.Run()

	return p.Decls, nil

}

func TestParseSingleVar(t *testing.T) {
	decls, err := readAndParse("tests/var.BUILD")
	if err != nil {
		t.Error(err)
	}
	decl := <-decls
	switch decl.(type) {
	case *ast.Assignment:
		asgn := decl.(*ast.Assignment)
		if asgn.Key != "UNDESIRED" {
			t.Fail()
		}
		switch asgn.Value.(type) {
		case *ast.BasicLit:
			val := asgn.Value.(*ast.BasicLit)
			if val.Kind != token.Quote || val.Value != "-fplan9-extensions" {
				t.Fail()
			}
		default:
			t.Fail()
		}
	default:
		t.Fail()
	}

}

func TestParseBoolVar(t *testing.T) {
	decls, err := readAndParse("tests/bool.BUILD")
	if err != nil {
		t.Error(err)
	}
	decl := <-decls
	switch decl.(type) {
	case *ast.Assignment:
		asgn := decl.(*ast.Assignment)
		if asgn.Key != "TRUE_BOOL" {
			t.Fail()
		}
		switch asgn.Value.(type) {
		case *ast.BasicLit:
			val := asgn.Value.(*ast.BasicLit)
			if val.Kind != token.True {
				t.Fail()
			}
		default:
			t.Fail()
		}
	default:
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

	decls, err := readAndParse("tests/slice.BUILD")
	if err != nil {
		t.Error(err)
	}
	decl := <-decls
	switch decl.(type) {
	case *ast.Assignment:
		asgn := decl.(*ast.Assignment)
		if asgn.Key != "C_FLAGS" {
			t.Log(asgn.Key)
			t.Fail()
		}
		switch asgn.Value.(type) {
		case *ast.Slice:
			val := asgn.Value.(*ast.Slice)
			for i, x := range val.Slice {
				if strs[i] != x.(*ast.BasicLit).Interface().(string) {
					t.Log(x.(string))
					t.Fail()
				}
			}
		default:
			t.Logf("not basic literal %T", asgn.Value)
			t.Fail()
		}
	default:
		t.Log("not an assignment")
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

	decls, err := readAndParse("tests/sliceWithOutLastComma.BUILD")
	if err != nil {
		t.Error(err)
	}
	decl := <-decls
	switch decl.(type) {
	case *ast.Assignment:
		asgn := decl.(*ast.Assignment)
		if asgn.Key != "C_FLAGS" {
			t.Log(asgn.Key)
			t.Fail()
		}
		switch asgn.Value.(type) {
		case *ast.Slice:
			val := asgn.Value.(*ast.Slice)
			for i, x := range val.Slice {
				if strs[i] != x.(*ast.BasicLit).Interface() {
					t.Log(x.(string))
					t.Fail()
				}
			}
		default:
			t.Logf("not basic literal %T", asgn.Value)
			t.Fail()
		}
	default:
		t.Log("not an assignment")
		t.Fail()
	}
}

func TestParseVarFunc(t *testing.T) {

	decls, err := readAndParse("tests/varFunc.BUILD")
	if err != nil {
		t.Error(err)
	}

	decl := <-decls
	switch decl.(type) {
	case *ast.Assignment:
		asgn := decl.(*ast.Assignment)
		v := asgn.Value
		switch v.(type) {
		case *ast.Func:
			f := v.(*ast.Func)
			if f.Name != "glob" {
				t.Fail()
			}
			q := f.AnonParams[0].(*ast.Slice)

			if q.Slice[0].(*ast.BasicLit).Interface() != "*.c" {
				t.Fail()
			}

		default:
			t.Fail()
		}

	default:
		t.Log("not an assignment")
		t.Fail()
	}

}

func TestParseAddition(t *testing.T) {

	decls, err := readAndParse("tests/addition.BUILD")
	if err != nil {
		t.Error(err)
	}

	decl := <-decls
	switch decl.(type) {
	case *ast.Assignment:
		v := decl.(*ast.Assignment).Value
		switch v.(type) {
		case *ast.Func:
			f := v.(*ast.Func)
			if f.Name != "addition" {
				t.Logf("%s is wrong function", f.Name)
				t.Fail()
			}

			if f.AnonParams[0].(*ast.Variable).Key != "CC_FLAGS" {
				t.Log("Was Expeting CC_FLAGS ")
				t.Fail()
			}

		default:
			t.Logf("was expectin a function not a %T", v)
			t.Fail()
		}
	default:
		t.Log("was expeting an assignment")
		t.Fail()
	}

	decl = <-decls
	switch decl.(type) {
	case *ast.Assignment:
		v := decl.(*ast.Assignment).Value
		switch v.(type) {
		case *ast.Func:
			f := v.(*ast.Func)
			if f.Name != "addition" {
				t.Logf("%s is wrong function", f.Name)
				t.Fail()
			}

			if f.AnonParams[0].(*ast.Variable).Key != "BB_FLAGS" {
				t.Log("was expecting BB_FLAGS")
				t.Fail()
			}

		default:
			t.Logf("was expectin a function not a %T", v)
			t.Fail()
		}
	default:
		t.Log("was expecting an assignment")
		t.Fail()
	}
}

func TestParseMap(t *testing.T) {
	decls, err := readAndParse("tests/map.BUILD")
	if err != nil {
		t.Error(err)
		return
	}

	decl := <-decls
	switch decl.(type) {
	case *ast.Assignment:
		v := decl.(*ast.Assignment).Value

		switch v.(type) {
		case map[string]interface{}:
			f := v.(map[string]interface{})
			if f["bla"] != "b" && f["foo"] != "p" {
				t.Fail()
			}
			return
		}
	}
}
func TestParseMapInFunc(t *testing.T) {
	decls, err := readAndParse("tests/mapinfunc.BUILD")
	if err != nil {
		t.Error(err)
		return
	}
	decl := <-decls

	switch decl.(type) {
	case *ast.Func:
		f := decl.(*ast.Func)
		if f.Params["exports"].(*ast.Map).Map["bla"].(*ast.BasicLit).Interface().(string) != "b" {
			t.Fail()
		}
		if f.Params["deps"].(*ast.Slice).Slice[0].(*ast.BasicLit).Interface() != ":libxstring" {
			t.Fail()
		}
		if f.Params["name"].(*ast.BasicLit).Interface().(string) != "test" {
			t.Fail()
		}
		if f.Params["srcs"].(*ast.Slice).Slice[0].(*ast.BasicLit).Interface() != "tests/test.c" {
			t.Fail()
		}
	default:
		t.Fail()
	}
}

func TestParseFunc(t *testing.T) {
	decls, err := readAndParse("tests/func.BUILD")
	if err != nil {
		t.Error(err)
		return
	}
	decl := <-decls

	switch decl.(type) {
	case *ast.Func:
		f := decl.(*ast.Func)
		if f.Params["copts"].(*ast.Variable).Key != "C_FLAGS" {
			t.Fail()
		}
		if f.Params["deps"].(*ast.Slice).Slice[0].(*ast.BasicLit).Interface() != ":libxstring" {
			t.Fail()
		}
		if f.Params["name"].(*ast.BasicLit).Interface().(string) != "test" {
			t.Fail()
		}
		if f.Params["srcs"].(*ast.Slice).Slice[0].(*ast.BasicLit).Interface() != "tests/test.c" {
			t.Fail()
		}
	default:
		t.Fail()
	}

}

func TestParseSmileyFunc(t *testing.T) {
	decls, err := readAndParse("tests/☺☹☻.BUILD")
	if err != nil {
		t.Error(err)
		return
	}
	decl := <-decls

	switch decl.(type) {
	case *ast.Func:
		f := decl.(*ast.Func)
		if f.Params["deps"].(*ast.Slice).Slice[0].(*ast.BasicLit).Interface() != ":☹☻☺" {
			t.Fail()
		}
		if f.Params["name"].(*ast.BasicLit).Interface().(string) != "☹☺☻" {
			t.Fail()
		}
		if f.Params["srcs"].(*ast.Slice).Slice[0].(*ast.BasicLit).Interface() != "☺☹☻.c" {
			t.Fail()
		}
	default:
		t.Fail()
	}
}

func TestParseSliceIndex(t *testing.T) {
	decls, err := readAndParse("tests/sliceIndex.BUILD")
	if err != nil {
		t.Error(err)
		return
	}
	decl := <-decls
	
	switch decl.(type) {
	case *ast.Assignment:

	default:
		t.Logf("%T", decl)
		t.Fail()
	}
	decl= <-decls
	switch decl.(type) {
	case *ast.Assignment:
		t.Log(prettyprint.AsJSON(decl.(*ast.Assignment)))
	case *ast.Error :
		t.Log(decl.(*ast.Error).Error)
	default:
		t.Logf("%T", decl)
		t.Fail()
	}
}

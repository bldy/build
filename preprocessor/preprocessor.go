// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package preprocessor // import "sevki.org/build/preprocessor"
import (
	"fmt"

	"sevki.org/build/ast"
)

type PreProcessor interface {
	Process(ast.Decl) (ast.Decl, error)
}

type DuplicateLoadChecker struct {
	Seen map[string]*ast.Func
}

func (dlc *DuplicateLoadChecker) Process(d ast.Decl) (ast.Decl, error) {
	switch d.(type) {
	case *ast.Func:
		f := d.(*ast.Func)
		if f.Name == "load" {
			file := f.AnonParams[0].(*ast.BasicLit).Value
			if exst, ok := dlc.Seen[file]; ok {
				dupeErr := `File %s is loaded more then once at these locations:
	 ./%s:%d: 
	 ./%s:%d: `

				return nil, fmt.Errorf(dupeErr, file, f.File, f.Start.Line, exst.File, exst.Start.Line)
			} else {
				dlc.Seen[file] = f
			}
		}
	}
	return d, nil
}

type DuplicateTargetNameChecker struct {
	Seen map[string]*ast.Func
}

func (dlc *DuplicateTargetNameChecker) Process(d ast.Decl) (ast.Decl, error) {

	switch d.(type) {
	case *ast.Func:
		f := d.(*ast.Func)
		if f.Name != "load" {
			file := f.Params["name"].(*ast.BasicLit).Value
			if exst, ok := dlc.Seen[file]; ok {
				dupeErr := `Target %s is declared more then once at these locations:
	 ./%s:%d: 
	 ./%s:%d: `

				return nil, fmt.Errorf(dupeErr, file, f.File, f.Start.Line, exst.File, exst.Start.Line)
			} else {
				dlc.Seen[file] = f
			}
		}
	}
	return d, nil
}

// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package preprocessor

import (
	"testing"

	"github.com/bldy/build/ast"
)

func TestDuplicateLoadChecker(t *testing.T) {
	dlc := &DuplicateLoadChecker{
		Seen: make(map[string]*ast.Func),
	}
	var params []interface{}
	params = append(params, &ast.BasicLit{Value: "filesomething"})
	dlc.Process(&ast.Func{
		Name:       "load",
		AnonParams: params,
	})
	_, err := dlc.Process(&ast.Func{
		Name:       "load",
		AnonParams: params,
	})
	if err == nil {
		t.Fail()
	}
}

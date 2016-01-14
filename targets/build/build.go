// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build // import "sevki.org/build/targets/build"

import (
	"log"

	"sevki.org/build/ast"
)

func init() {
	if err := ast.Register("gen_rule", GenRule{}); err != nil {
		log.Fatal(err)
	}
	if err := ast.Register("group", Group{}); err != nil {
		log.Fatal(err)
	}
}

// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harvey // import "sevki.org/build/targets/harvey"

import (
	"log"

	"sevki.org/build/ast"
)

func init() {
	if err := ast.Register("mk_sys", MkSys{}); err != nil {
		log.Fatal(err)
	}
	if err := ast.Register("group", Group{}); err != nil {
		log.Fatal(err)
	}
	if err := ast.Register("man_page", ManPage{}); err != nil {
		log.Fatal(err)
	}
}

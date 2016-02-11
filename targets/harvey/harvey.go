// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harvey // import "sevki.org/build/targets/harvey"

import (
	"log"

	"sevki.org/build/ast"
)

func init() {
	if err := ast.Register("kernel", Kernel{}); err != nil {
		log.Fatal(err)
	}
	if err := ast.Register("config", Config{}); err != nil {
		log.Fatal(err)
	}
	if err := ast.Register("usb", USB{}); err != nil {
		log.Fatal(err)
	}
	if err := ast.Register("sed", Sed{}); err != nil {
		log.Fatal(err)
	}
	if err := ast.Register("strip", Strip{}); err != nil {
		log.Fatal(err)
	}
	if err := ast.Register("objcopy", ObjCopy{}); err != nil {
		log.Fatal(err)
	}
	if err := ast.Register("mk_sys", MkSys{}); err != nil {
		log.Fatal(err)
	}
	if err := ast.Register("elf_to_c", ElfToC{}); err != nil {
		log.Fatal(err)
	}

	if err := ast.Register("data_to_c", DataToC{}); err != nil {
		log.Fatal(err)
	}
}

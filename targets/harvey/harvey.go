// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harvey // import "sevki.org/build/targets/harvey"

import (
	"log"

	"sevki.org/build/internal"
)

func init() {
	if err := internal.Register("kernel", Kernel{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("config", Config{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("usb", USB{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("sed", Sed{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("strip", Strip{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("objcopy", ObjCopy{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("mk_sys", MkSys{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("elf_to_c", ElfToC{}); err != nil {
		log.Fatal(err)
	}

	if err := internal.Register("data_to_c", DataToC{}); err != nil {
		log.Fatal(err)
	}
}

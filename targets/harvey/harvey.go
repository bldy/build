// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harvey

import (
	"log"

	"bldy.build/build/internal"
)

func init() {
	if err := internal.Register("move", Move{}); err != nil {
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
	if err := internal.Register("qemu", Qemu{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("old_build", OldBuild{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("data_to_c", DataToC{}); err != nil {
		log.Fatal(err)
	}
}

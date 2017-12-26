// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build

import (
	"log"

	"bldy.build/build/internal"
)

func init() {
	if err := internal.Register("gen_rule", GenRule{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("group", Group{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("template", Template{}); err != nil {
		log.Fatal(err)
	}
}

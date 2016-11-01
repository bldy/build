// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build // import "sevki.org/build/targets/build"

import (
	"log"

	"sevki.org/build/internal"
)

func init() {
	if err := internal.Register("gen_rule", GenRule{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("group", Group{}); err != nil {
		log.Fatal(err)
	}
}

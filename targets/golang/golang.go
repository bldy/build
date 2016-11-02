// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package golang

import (
	"log"
	"os/exec"
	"strings"

	"github.com/bldy/build/internal"
)

var (
	gover string
)

func init() {
	if err := internal.Register("go_build", GoBuild{}); err != nil {
		log.Fatal(err)
	}

	if out, err := exec.Command(Compiler(), "--version").Output(); err != nil {
		gover = "deadbeef"
	} else {
		gover = strings.TrimSpace(string(out))
	}
}

func Compiler() string {
	return "go"

}

// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "sevki.org/build/targets/cc"

import (
	"fmt"
	"log"
	"os/exec"

	"strings"

	"os"

	"sevki.org/build/ast"
)

var CCVersion = ""
var cc = ""

func init() {

	if cc = os.Getenv("CC"); cc == "" {
		cc = "CC"
	}

	if out, err := exec.Command(cc, "--version").Output(); err != nil {
		CCVersion = "deadbeef"
	} else {
		CCVersion = strings.TrimSpace(string(out))
	}
	if err := ast.Register("cc_library", CLib{}); err != nil {
		log.Fatal(err)
	}
	if err := ast.Register("cxx_library", CLib{}); err != nil {
		log.Fatal(err)
	}
	if err := ast.Register("cc_binary", CBin{}); err != nil {
		log.Fatal(err)
	}
}

func compiler() string {
	if tpfx := os.Getenv("TOOLPREFIX"); tpfx == "" {
		return cc
	} else {
		return fmt.Sprintf("%s%s", tpfx, cc)
	}
}
func ar() string {
	if tpfx := os.Getenv("TOOLPREFIX"); tpfx == "" {
		return "ar"
	} else {
		return fmt.Sprintf("%s%s", tpfx, "ar")
	}
}
func ld() string {
	if tpfx := os.Getenv("TOOLPREFIX"); tpfx == "" {
		return "ld"
	} else {
		return fmt.Sprintf("%s%s", tpfx, "ld")
	}
}

type Sources []string

type CompilerFlags []string

type Includes []string

func (s Sources) String() string {
	return strings.Join(s, " ")
}

func (s Includes) Includes() (incs []string) {
	for _, i := range s {
		incs = append(incs, fmt.Sprintf("-I%s", i))
	}
	return
}

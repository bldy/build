// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc

import (
	"fmt"
	"log"
	"os/exec"

	"strings"

	"os"

	"bldy.build/build/internal"
	"bldy.build/build/project"
)

var (
	CCVersion = ""
	cc        = ""
	ld        = ""
	ar        = ""
	CCENV     = os.Environ()
)

func init() {

	CCENV = append(CCENV, fmt.Sprintf("%s=%s", "C_INCLUDE_PATH", "include"))
	CCENV = append(CCENV, fmt.Sprintf("%s=%s", "LIBRARY_PATH", "lib"))

	if cc = project.Getenv("CC"); cc == "" {
		cc = "CC"
	}
	if ld = project.Getenv("LD"); ld == "" {
		ld = "ld"
	}
	if ar = project.Getenv("AR"); ar == "" {
		ar = "ar"
	}
	if out, err := exec.Command(Compiler(), "--version").Output(); err != nil {
		CCVersion = "deadbeef"
	} else {
		CCVersion = strings.TrimSpace(string(out))
	}
	if err := internal.Register("cc_library", CLib{}); err != nil {
		log.Fatal(err)
	}
	if err := internal.Register("cc_binary", CBin{}); err != nil {
		log.Fatal(err)
	}
}

func Compiler() string {
	if tpfx := project.Getenv("TOOLPREFIX"); tpfx == "" {
		return cc
	} else {
		return fmt.Sprintf("%s%s", tpfx, cc)
	}
}

func Archiver() string {
	if tpfx := project.Getenv("TOOLPREFIX"); tpfx == "" {
		return ar
	} else {
		return fmt.Sprintf("%s%s", tpfx, ar)
	}
}
func Linker() string {
	if tpfx := project.Getenv("TOOLPREFIX"); tpfx == "" {
		return ld
	} else {
		return fmt.Sprintf("%s%s", tpfx, ld)
	}
}

// Stripper returns the strip binary
func Stripper() string {
	if tpfx := project.Getenv("TOOLPREFIX"); tpfx == "" {
		return "strip"
	} else {
		return fmt.Sprintf("%s%s", tpfx, "strip")
	}
}

func includes(s []string) (incs []string) {
	for _, i := range s {
		incs = append(incs, fmt.Sprintf("-I%s", i))
	}
	return
}

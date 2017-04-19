// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo

package cc

import (
	"crypto/sha1"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"bldy.build/build"
	"bldy.build/build/racy"
	"sevki.org/lib/prettyprint"
)

type CTest struct {
	Name            string        `cxx_test:"name" cc_test:"name"`
	Sources         []string      `cxx_test:"srcs" cc_test:"srcs" build:"path"`
	Dependencies    []string      `cxx_test:"deps" cc_test:"deps"`
	Includes        Includes      `cxx_test:"headers" cc_test:"includes" build:"path"`
	Headers         []string      `cxx_test:"exported_headers" cc_test:"hdrs" build:"path"`
	CompilerOptions CompilerFlags `cxx_test:"compiler_flags" cc_test:"copts"`
	LinkerOptions   []string      `cxx_test:"linker_flags" cc_test:"linkopts"`
	LinkerFile      string        `cxx_test:"ld" cc_test:"ld" build:"path"`
	Static          bool          `cxx_test:"linkstatic" cc_test:"linkstatic"`
	Strip           bool          `cxx_test:"strip" cc_test:"strip"`
	AlwaysLink      bool          `cxx_test:"alwayslink" cc_test:"alwayslink"`
}

func (ct *CTest) Hash() []byte {

	h := sha1.New()
	io.WriteString(h, CCVersion)
	io.WriteString(h, ct.Name)
	racy.HashFilesWithExt(h, []string(ct.Includes), ".h")
	racy.HashFilesWithExt(h, ct.Sources, ".c")
	racy.HashStrings(h, ct.CompilerOptions)
	racy.HashStrings(h, ct.LinkerOptions)
	return h.Sum(nil)
}

func (ct *CTest) Build(c *build.Runner) error {
	c.Println(prettyprint.AsJSON(ct))
	params := []string{"-c"}
	params = append(params, ct.CompilerOptions...)
	params = append(params, ct.Sources...)

	params = append(params, ct.Includes.Includes()...)

	if err := c.Exec(Compiler(), CCENV, params); err != nil {
		return fmt.Errorf(err.Error())
	}

	ldparams := []string{"-o", ct.Name}
	ldparams = append(ldparams, ct.LinkerOptions...)
	if ct.LinkerFile != "" {
		ldparams = append(ldparams, ct.LinkerFile)
	}

	// This is done under the assumption that each src file put in this thing
	// here will comeout as a .o file
	for _, f := range ct.Sources {
		_, fname := filepath.Split(f)
		ldparams = append(ldparams, fmt.Sprintf("%s.o", fname[:strings.LastIndex(fname, ".")]))
	}

	haslib := false
	for _, dep := range ct.Dependencies {
		d := split(dep, ":")
		if len(d) < 3 {
			continue
		}
		if d[:3] == "lib" {
			if ct.AlwaysLink {
				ldparams = append(ldparams, fmt.Sprintf("%s.a", d))
			} else {
				if !haslib {
					ldparams = append(ldparams, "-L", "lib")
					haslib = true
				}
				ldparams = append(ldparams, fmt.Sprintf("-l%s", d[3:]))
			}
		}

		// kernel specific
		if len(d) < 4 {
			continue
		}
		if d[:4] == "klib" {
			ldparams = append(ldparams, fmt.Sprintf("%s.a", d))
		}
	}

	if err := c.Exec(Linker(), CCENV, ldparams); err != nil {
		return fmt.Errorf(err.Error())
	}
	if ct.Strip {
		sparams := []string{"-o", ct.Name, ct.Name}
		if err := c.Exec(Stripper(), nil, sparams); err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	return nil
}

func (ct *CTest) Installs() map[string]string {
	exports := make(map[string]string)

	exports[filepath.Join("bin", ct.Name)] = ct.Name

	return exports
}

func (ct *CTest) GetName() string {
	return ct.Name
}
func (ct *CTest) GetDependencies() []string {
	return ct.Dependencies
}

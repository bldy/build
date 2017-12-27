// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc

import (
	"fmt"
	"io"
	"strings"

	"bldy.build/build/executor"
	"bldy.build/build/racy"

	"path/filepath"
)

type CLib struct {
	Name            string        `cxx_library:"name" cc_library:"name"`
	Sources         []string      `cxx_library:"srcs" cc_library:"srcs" build:"path" ext:".c,.S,.cpp,.cc"`
	Dependencies    []string      `cxx_library:"deps" cc_library:"deps"`
	Includes        Includes      `cxx_library:"headers" cc_library:"includes" build:"path" ext:".h,.c,.S"`
	Headers         []string      `cxx_library:"exported_headers" cc_library:"hdrs" build:"path" ext:".h,.c,.S"`
	CompilerOptions CompilerFlags `cxx_library:"compiler_flags" cc_library:"copts"`
	LinkerOptions   []string      `cxx_library:"linker_flags" cc_library:"linkopts"`
	LinkStatic      bool          `cxx_library:"linkstatic" cc_library:"linkstatic"`
	AlwaysLink      bool          `cxx_library:"alwayslink" cc_library:"alwayslink"`
}

func (cl *CLib) Hash() []byte {
	h := racy.New()

	io.WriteString(h, CCVersion)
	io.WriteString(h, cl.Name)
	io.WriteString(h, "clib")
	racy.HashStrings(h, cl.CompilerOptions)
	racy.HashStrings(h, cl.LinkerOptions)
	if cl.LinkStatic {
		io.WriteString(h, "static")
	}
	return racy.XOR(
		h.Sum(nil),
		racy.HashFilesForExt([]string(cl.Includes), ".h"),
		racy.HashFilesForExt(cl.Sources, ".c"),
		racy.HashFilesForExt(cl.Sources, ".S"),
	)
}

func (cl *CLib) Build(e *executor.Executor) error {
	params := []string{"-c"}
	params = append(params, cl.CompilerOptions...)
	params = append(params, cl.LinkerOptions...)
	params = append(params, cl.Includes.Includes()...)
	params = append(params, cl.Sources...)

	if err := e.Exec(Compiler(), CCENV, params); err != nil {
		return fmt.Errorf(err.Error())
	}

	libName := fmt.Sprintf("%s.a", cl.Name)
	params = []string{"-rs", libName}
	params = append(params, cl.LinkerOptions...)
	// This is done under the assumption that each src file put in this thing
	// here will comeout as a .o file
	for _, f := range cl.Sources {
		_, filename := filepath.Split(f)
		params = append(params, fmt.Sprintf("%s.o", filename[:strings.LastIndex(filename, ".")]))
	}

	if err := e.Exec(Archiver(), CCENV, params); err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}
func (cl *CLib) Installs() map[string]string {
	exports := make(map[string]string)
	libName := fmt.Sprintf("%s.a", cl.Name)
	if cl.AlwaysLink {
		exports[libName] = libName
	} else {
		exports[filepath.Join("lib", libName)] = libName
	}
	return exports
}
func (cl *CLib) GetName() string {
	return cl.Name
}

func (cl *CLib) GetDependencies() []string {
	return cl.Dependencies
}

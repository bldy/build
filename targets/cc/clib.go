// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "sevki.org/build/targets/cc"

import (
	"bytes"
	"crypto/sha1"
	"io"

	"strings"

	"fmt"

	"sevki.org/build/util"

	"path/filepath"

	"sevki.org/build"
)

type CLib struct {
	Name            string        `cxx_library:"name" cc_library:"name"`
	Sources         []string      `cxx_library:"srcs" cc_library:"srcs" build:"path"`
	Dependencies    []string      `cxx_library:"deps" cc_library:"deps"`
	Includes        Includes      `cxx_library:"headers" cc_library:"includes" build:"path"`
	Headers         []string      `cxx_library:"exported_headers" cc_library:"hdrs" build:"path"`
	CompilerOptions CompilerFlags `cxx_library:"compiler_flags" cc_library:"copts"`
	LinkerOptions   []string      `cxx_library:"linker_flags" cc_library:"linkopts"`
	LinkShared      bool
	LinkStatic      bool
	Source          string
	buf             bytes.Buffer
}

func (cl *CLib) Hash() []byte {
	h := sha1.New()
	io.WriteString(h, CCVersion)
	io.WriteString(h, cl.Name)
	util.HashFiles(h, cl.Includes)
	io.WriteString(h, "clib")
	util.HashFiles(h, []string(cl.Sources))
	util.HashStrings(h, cl.CompilerOptions)
	util.HashStrings(h, cl.LinkerOptions)
	if cl.LinkShared {
		io.WriteString(h, "shared")
	}
	if cl.LinkStatic {
		io.WriteString(h, "static")
	}
	return h.Sum(nil)
}

func (cl *CLib) Build(c *build.Context) error {
	params := []string{}
	params = append(params, cl.CompilerOptions...)
	params = append(params, cl.LinkerOptions...)
	params = append(params, cl.Sources...)
	params = append(params, cl.Includes.Includes()...)

	c.Println(strings.Join(append([]string{compiler()}, params...), " "))

	if err := c.Exec(compiler(), CCENV, params); err != nil {
		c.Println(err.Error())
		return fmt.Errorf(cl.buf.String())
	}

	libName := fmt.Sprintf("%s.a", cl.Name)
	params = []string{"-rs", libName}

	// This is done under the assumption that each src file put in this thing
	// here will comeout as a .o file
	for _, f := range cl.Sources {
		_, filename := filepath.Split(f)
		params = append(params, fmt.Sprintf("%s.o", filename[:strings.LastIndex(filename, ".")]))
	}

	c.Println(strings.Join(append([]string{ar()}, params...), " "))
	if err := c.Exec(ar(), CCENV, params); err != nil {
		c.Println(err.Error())
		return fmt.Errorf(cl.buf.String())
	}

	return nil
}
func (cl *CLib) Installs() map[string]string {

	exports := make(map[string]string)
	exports[fmt.Sprintf("%s.a", cl.Name)] = "lib"
	return exports
}
func (cl *CLib) GetName() string {
	return cl.Name
}

func (cl *CLib) GetDependencies() []string {
	return cl.Dependencies
}
func (cl *CLib) GetSource() string {
	return cl.Source
}

func (cl *CLib) Reader() io.Reader {
	return &cl.buf
}

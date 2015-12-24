package cc

import (
	"bytes"
	"crypto/sha1"
	"fmt"

	"io"
	"strings"

	"path/filepath"

	"sevki.org/build/build"
	"sevki.org/build/util"
)

type CBin struct {
	Name            string        `cxx_binary:"name" cc_binary:"name"`
	Sources         Sources       `cxx_binary:"srcs" cc_binary:"srcs" build:"path"`
	Dependencies    []string      `cxx_binary:"deps" cc_binary:"deps"`
	Includes        Includes      `cxx_binary:"headers" cc_binary:"includes" build:"path"`
	Headers         []string      `cxx_binary:"exported_headers" cc_binary:"hdrs" build:"path"`
	CompilerOptions CompilerFlags `cxx_binary:"compiler_flags" cc_binary:"copts"`
	LinkerOptions   []string      `cxx_binary:"linker_flags" cc_binary:"linkopts"`
	LinkShared      bool
	LinkStatic      bool
	Source          string
	buf             bytes.Buffer
}

func split(s string, c string) string {
	i := strings.Index(s, c)
	if i < 0 {
		return s
	}

	return s[i+1:]
}
func (cb *CBin) Hash() []byte {
	for _, dep := range cb.Dependencies {
		d := split(dep, ":")

		if strings.TrimLeft(d, "lib") != d {
			cb.LinkerOptions = append(cb.LinkerOptions, fmt.Sprintf("-l%s", d[3:]))
		}
	}

	h := sha1.New()
	io.WriteString(h, CCVersion)
	io.WriteString(h, cb.Name)
	util.HashFiles(h, cb.Includes)
	util.HashFiles(h, []string(cb.Sources))
	util.HashStrings(h, cb.CompilerOptions)
	util.HashStrings(h, cb.LinkerOptions)
	if cb.LinkShared {
		io.WriteString(h, "shared")
	}
	if cb.LinkStatic {
		io.WriteString(h, "static")
	}
	return h.Sum(nil)
}

func (cb *CBin) Build(c *build.Context) error {

	params := []string{}
	params = append(params, cb.CompilerOptions...)
	params = append(params, cb.Sources...)
	params = append(params, cb.Includes.Includes()...)

	c.Println(strings.Join(append([]string{compiler()}, params...), " "))

	if err := c.Exec(compiler(), nil, params); err != nil {
		c.Println(err.Error())
		return fmt.Errorf(cb.buf.String())
	}

	params = []string{"-rs", cb.Name}

	// This is done under the assumption that each src file put in this thing
	// here will comeout as a .o file
	for _, f := range cb.Sources {
		_, filename := filepath.Split(f)
		ext := filepath.Ext(filename)
		params = append(params, fmt.Sprintf("%s.o", strings.TrimRight(filename, ext)))
	}
	params = append(params, "-L", "lib")
	params = append(params, cb.LinkerOptions...)

	c.Println(strings.Join(append([]string{ld()}, params...), " "))
	if err := c.Exec(ld(), nil, params); err != nil {
		c.Println(err.Error())
		return fmt.Errorf(cb.buf.String())
	}

	return nil
}

func (cb *CBin) Installs() map[string]string {
	exports := make(map[string]string)
	exports[cb.Name] = ""
	return exports
}

func (cb *CBin) GetName() string {
	return cb.Name
}
func (cb *CBin) GetDependencies() []string {
	return cb.Dependencies
}

func (cb *CBin) GetSource() string {
	return cb.Source
}

func (cl *CBin) Reader() io.Reader {
	return &cl.buf
}

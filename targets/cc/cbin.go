package cc

import (
	"crypto/sha1"
	"fmt"

	"io"
	"strings"

	"path/filepath"

	"sevki.org/build"
	"sevki.org/build/util"
)

type CBin struct {
	Name            string        `cxx_binary:"name" cc_binary:"name"`
	Sources         []string      `cxx_binary:"srcs" cc_binary:"srcs" build:"path"`
	Dependencies    []string      `cxx_binary:"deps" cc_binary:"deps"`
	Includes        Includes      `cxx_binary:"headers" cc_binary:"includes" build:"path"`
	Headers         []string      `cxx_binary:"exported_headers" cc_binary:"hdrs" build:"path"`
	CompilerOptions CompilerFlags `cxx_binary:"compiler_flags" cc_binary:"copts"`
	LinkerOptions   []string      `cxx_binary:"linker_flags" cc_binary:"linkopts"`
	Source          string
}

func split(s string, c string) string {
	i := strings.Index(s, c)
	if i < 0 {
		return s
	}

	return s[i+1:]
}
func (cb *CBin) Hash() []byte {

	h := sha1.New()
	io.WriteString(h, CCVersion)
	io.WriteString(h, cb.Name)
	util.HashFiles(h, cb.Includes)
	util.HashFiles(h, []string(cb.Sources))
	util.HashStrings(h, cb.CompilerOptions)
	util.HashStrings(h, cb.LinkerOptions)
	return h.Sum(nil)
}

func (cb *CBin) Build(c *build.Context) error {

	params := []string{"-v"}
	params = append(params, cb.CompilerOptions...)
	params = append(params, cb.Sources...)

	params = append(params, cb.Includes.Includes()...)

	c.Println(strings.Join(append([]string{compiler()}, params...), " "))

	if err := c.Exec(compiler(), CCENV, params); err != nil {
		return fmt.Errorf(err.Error())
	}

	params = []string{"-o", cb.Name}

	// This is done under the assumption that each src file put in this thing
	// here will comeout as a .o file
	for _, f := range cb.Sources {
		_, fname := filepath.Split(f)
		params = append(params, fmt.Sprintf("%s.o", fname[:strings.LastIndex(fname, ".")]))
	}
	cb.LinkerOptions = append(cb.LinkerOptions, "-L", "lib")

	for _, dep := range cb.Dependencies {
		d := split(dep, ":")

		if strings.TrimLeft(d, "lib") != d {
			cb.LinkerOptions = append(cb.LinkerOptions, fmt.Sprintf("-l%s", d[3:]))
		}
	}

	params = append(params, cb.LinkerOptions...)

	c.Println(CCENV)
	c.Println(strings.Join(append([]string{ld()}, params...), " "))
	if err := c.Exec(ld(), CCENV, params); err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

func (cb *CBin) Installs() map[string]string {
	exports := make(map[string]string)
	exports[cb.Name] = "bin"
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

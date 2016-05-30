package cc

import (
	"crypto/sha1"
	"fmt"

	"io"
	"strings"

	"path/filepath"

	"sevki.org/build"
	"sevki.org/build/util"
	"sevki.org/lib/prettyprint"
)

type CBin struct {
	Name            string        `cxx_binary:"name" cc_binary:"name"`
	Sources         []string      `cxx_binary:"srcs" cc_binary:"srcs" build:"path"`
	Dependencies    []string      `cxx_binary:"deps" cc_binary:"deps"`
	Includes        Includes      `cxx_binary:"headers" cc_binary:"includes" build:"path"`
	Headers         []string      `cxx_binary:"exported_headers" cc_binary:"hdrs" build:"path"`
	CompilerOptions CompilerFlags `cxx_binary:"compiler_flags" cc_binary:"copts"`
	LinkerOptions   []string      `cxx_binary:"linker_flags" cc_binary:"linkopts"`
	LinkerFile      string        `cxx_binary:"ld" cc_binary:"ld" build:"path"`
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
	util.HashFilesWithExt(h, cb.Includes, ".h")
	util.HashFiles(h, []string(cb.Sources))
	util.HashStrings(h, cb.CompilerOptions)
	util.HashStrings(h, cb.LinkerOptions)
	return h.Sum(nil)
}

func (cb *CBin) Build(c *build.Context) error {
	c.Println(prettyprint.AsJSON(cb))
	params := []string{"-c"}
	params = append(params, cb.CompilerOptions...)
	params = append(params, cb.Sources...)

	params = append(params, cb.Includes.Includes()...)

	if err := c.Exec(Compiler(),CCENV, params); err != nil {
		return fmt.Errorf(err.Error())
	}

	ldparams := []string{"-o", cb.Name}
	ldparams = append(ldparams, cb.LinkerOptions...)
	if cb.LinkerFile != "" {
		ldparams = append(ldparams, cb.LinkerFile)
	}
	// This is done under the assumption that each src file put in this thing
	// here will comeout as a .o file
	for _, f := range cb.Sources {
		_, fname := filepath.Split(f)
		ldparams = append(ldparams, fmt.Sprintf("%s.o", fname[:strings.LastIndex(fname, ".")]))
	}

	ldparams = append(ldparams, "-L", "lib")

	for _, dep := range cb.Dependencies {
		d := split(dep, ":")
		if len(d) < 3 {
			continue
		}
		if d[:3] == "lib" {
			ldparams = append(ldparams, fmt.Sprintf("-l%s", d[3:]))
		}
	}

	if err := c.Exec(Linker(), CCENV, ldparams); err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

func (cb *CBin) Installs() map[string]string {
	exports := make(map[string]string)

	exports[filepath.Join("bin", cb.Name)] = cb.Name

	return exports
}

func (cb *CBin) GetName() string {
	return cb.Name
}
func (cb *CBin) GetDependencies() []string {
	return cb.Dependencies
}

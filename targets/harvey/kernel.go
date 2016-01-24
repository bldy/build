package harvey

import (
	"crypto/sha1"
	"fmt"

	"io"
	"strings"

	"path/filepath"

	"sevki.org/build"
	"sevki.org/build/targets/cc"
	"sevki.org/build/util"
	"sevki.org/lib/prettyprint"
)

type Kernel struct {
	Name            string           `kernel:"name"`
	Sources         []string         `kernel:"srcs" build:"path"`
	Dependencies    []string         `kernel:"deps"`
	Includes        cc.Includes      `kernel:"includes" build:"path"`
	Headers         []string         `kernel:"hdrs" build:"path"`
	CompilerOptions cc.CompilerFlags `kernel:"copts"`
	LinkerOptions   []string         `kernel:"linkopts"`
	LinkerFile      string           `kernel:"ld" build:"path"`
}

func split(s string, c string) string {
	i := strings.Index(s, c)
	if i < 0 {
		return s
	}

	return s[i+1:]
}
func (k *Kernel) Hash() []byte {

	h := sha1.New()
	io.WriteString(h, cc.CCVersion)
	io.WriteString(h, k.Name)
	util.HashFiles(h, k.Includes)
	util.HashFiles(h, []string(k.Sources))
	util.HashStrings(h, k.CompilerOptions)
	util.HashStrings(h, k.LinkerOptions)
	return h.Sum(nil)
}

func (k *Kernel) Build(c *build.Context) error {
	c.Println(prettyprint.AsJSON(k))
	params := []string{"-c"}
	params = append(params, k.CompilerOptions...)
	params = append(params, k.Sources...)

	params = append(params, k.Includes.Includes()...)

	if err := c.Exec(cc.Compiler(), cc.CCENV, params); err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

func (k *Kernel) Installs() map[string]string {
	exports := make(map[string]string)

	exports[filepath.Join("bin", k.Name)] = k.Name

	return exports
}

func (k *Kernel) GetName() string {
	return k.Name
}
func (k *Kernel) GetDependencies() []string {
	return k.Dependencies
}

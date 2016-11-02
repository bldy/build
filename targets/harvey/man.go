package harvey

import (
	"crypto/sha1"
	"fmt"
	"io"
	"path/filepath"

	"strings"

	"github.com/bldy/build"
	"github.com/bldy/build/util"
)
import "os"

type ManPage struct {
	Name         string   `man_page:"name"`
	Dependencies []string `man_page:"deps"`
	Sources      []string `man_page:"srcs" build:"path"`
}

func (mp *ManPage) Hash() []byte {
	h := sha1.New()

	io.WriteString(h, mp.Name)

	util.HashFiles(h, mp.Sources)
	util.HashStrings(h, os.Environ())
	return []byte{}
}

func (mp *ManPage) Build(c *build.Context) error {
	for _, m := range mp.Sources {
		params := []string{"<"}
		params = append(params, m)

		params = append(params, ">")
		params = append(params, fmt.Sprintf("%s.html", m))

		c.Println(strings.Join(append([]string{"man2html"}, params...), " "))

		if err := c.Exec("man2html", nil, params); err != nil {
			return err
		}
	}
	return nil
}

func (mp *ManPage) GetName() string {
	return mp.Name
}

func (mp *ManPage) GetDependencies() []string {
	return mp.Dependencies
}
func (mp *ManPage) Installs() map[string]string {
	exports := make(map[string]string)
	for _, m := range mp.Sources {
		_, f := filepath.Split(m)
		exports[f] = "man"
	}

	return exports
}

package harvey

import (
	"crypto/sha1"
	"fmt"

	"bldy.build/build/url"
	"bldy.build/build/util"

	"io"

	"bldy.build/build"
)

type Move struct {
	Name         string            `move:"name"`
	Dependencies []string          `move:"deps"`
	Exports      map[string]string `move:"installs" build:"expand"`
}

func (m *Move) Hash() []byte {
	h := sha1.New()
	io.WriteString(h, m.Name)
	return h.Sum(nil)
}

func (m *Move) Build(c *build.Context) error {
	return nil
}

func (m *Move) Installs() map[string]string {
	installs := make(map[string]string)
	for _, d := range m.Dependencies {
		u := url.Parse(d)

		// Lest figure out where things go
		// first the libraries
		if len(u.Target) > 3 && u.Target[:3] == "lib" {
			installs[fmt.Sprintf("%s/lib/%s.a", util.Getenv("ARCH"), u.Target)] = fmt.Sprintf("lib/%s.a", u.Target)
			continue
		}
		// cmds next
		a := "sys/src/cmd"
		if u.Package == a {
			installs[fmt.Sprintf("%s/bin/%s", util.Getenv("ARCH"), u.Target)] = fmt.Sprintf("bin/%s", u.Target)
		}
	}
	for k, v := range m.Exports {
		installs[k] = v
	}
	return installs
}

func (m *Move) GetName() string {
	return m.Name
}
func (m *Move) GetDependencies() []string {
	return m.Dependencies
}

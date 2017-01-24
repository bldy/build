package build

import (
	"crypto/sha1"
	"io"

	"strings"

	"os"

	"bldy.build/build"
	"bldy.build/build/racy"
)

type GenRule struct {
	Name         string   `gen_rule:"name"`
	Dependencies []string `gen_rule:"deps"`
	Commands     []string `gen_rule:"cmds"`
}

func (g *GenRule) Hash() []byte {
	h := sha1.New()

	io.WriteString(h, g.Name)
	racy.HashStrings(h, g.Commands)
	racy.HashStrings(h, os.Environ())
	return []byte{}
}

func (g *GenRule) Build(c *build.Context) error {
	for _, cmd := range g.Commands {
		strs := strings.Split(cmd, " ")

		if err := c.Exec(strs[0], nil, strs[1:]); err != nil {
			c.Println(err.Error())
			return err
		}
	}
	return nil
}

func (g *GenRule) GetName() string {
	return g.Name
}

func (g *GenRule) GetDependencies() []string {
	return g.Dependencies
}
func (g *GenRule) Installs() map[string]string {
	return nil
}

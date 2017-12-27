package build

import (
	"io"

	"strings"

	"os"

	"bldy.build/build/executor"
	"bldy.build/build/racy"
)

type GenRule struct {
	Name         string   `gen_rule:"name"`
	Dependencies []string `gen_rule:"deps"`
	Commands     []string `gen_rule:"cmds"`
}

func (g *GenRule) Hash() []byte {
	h := racy.New()

	io.WriteString(h, g.Name)
	racy.HashStrings(h, g.Commands)
	racy.HashStrings(h, os.Environ())
	return []byte{}
}

func (g *GenRule) Build(e *executor.Executor) error {
	for _, cmd := range g.Commands {
		strs := strings.Split(cmd, " ")

		if err := e.Exec(strs[0], nil, strs[1:]); err != nil {
			e.Println(err.Error())
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

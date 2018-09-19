package golang

import (
	"bldy.build/build/executor"
	"bldy.build/build/label"
)

type Go struct {
	name         string        `group:"name"`
	dependencies []label.Label `group:"deps"`
}

func (g *Go) Name() string {
	panic("not implemented")
}

func (g *Go) Dependencies() []label.Label {
	panic("not implemented")
}

func (g *Go) Outputs() []string {
	panic("not implemented")
}

func (g *Go) Hash() []byte {
	panic("not implemented")
}

func (g *Go) Build(*executor.Executor) error {
	panic("not implemented")
}

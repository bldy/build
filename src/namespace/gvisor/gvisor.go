package gvisor

import (
	"context"
	"os"

	"bldy.build/bldy/src/namespace"

)

func New() (namespace.Namespace, error) { return &gvisor{}, nil }

type gvisor struct{}

func (g *gvisor) Bind(new string, old string, flags int) {
	panic("not implemented")
}

func (g *gvisor) Mount(new string, old string, flags int) {
	panic("not implemented")
}

func (g *gvisor) Cmd(ctx context.Context, cmd string, args ...string) namespace.Cmd {
	panic("not implemented")
}

func (g *gvisor) Mkdir(name string) error {
	panic("not implemented")
}

func (g *gvisor) Open(name string) (*os.File, error) {
	panic("not implemented")
}

func (g *gvisor) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	panic("not implemented")
}

func (g *gvisor) Create(name string) (*os.File, error) {
	panic("not implemented")
}

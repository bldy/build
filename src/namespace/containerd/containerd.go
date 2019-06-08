package containerd

// https://github.com/containerd/containerd/blob/master/PLUGINS.md

import (
	"context"
	"os"

	"bldy.build/bldy/src/namespace"

	"github.com/containerd/containerd"
)

var client = func() *containerd.Client {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		panic(err)
	}
	return client
}()

func New() (namespace.Namespace, error) {
	return &contarinerdNS{}, nil

}

type contarinerdNS struct{}

func (c *contarinerdNS) Bind(new string, old string, flags int) {
	panic("not implemented")
}

func (c *contarinerdNS) Mount(new string, old string, flags int) {
	panic("not implemented")
}

func (c *contarinerdNS) Cmd(ctx context.Context, cmd string, args ...string) namespace.Cmd {
	panic("not implemented")
}

func (c *contarinerdNS) Mkdir(name string) error {
	panic("not implemented")
}

func (c *contarinerdNS) Open(name string) (*os.File, error) {
	panic("not implemented")
}

func (c *contarinerdNS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	panic("not implemented")
}

func (c *contarinerdNS) Create(name string) (*os.File, error) {
	panic("not implemented")
}
func MountWorkspace(s string) {
	panic("not implemented")
}

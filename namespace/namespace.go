package namespace

import (
	"context"
	"os"
)

const (
	MREPL int = iota
	MBEFORE
	MAFTER
)

type Namespace interface {
	Bind(new, old string, flags int)
	Mount(new, old string, flags int)
	Cmd(ctx context.Context, cmd string, args ...string) Cmd
	Mkdir(name string) error
	Open(name string) (*os.File, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Create(name string) (*os.File, error)
}

type Workspace interface {
	Namespace
	MountWorkspace(s string)
}

type Cmd interface {
	Run() error
	Start() error
	Wait() error
	CombinedOutput() ([]byte, error)
	Output() ([]byte, error)
}
// Package ziggy is the is the VM for stardust scripts which are
// derivied from starlark scripts
package ziggy

import (
	"io"
	"log"

	"golang.org/x/exp/errors/fmt"

	"bldy.build/bldy/src/build"
	"bldy.build/bldy/src/url"
	"go.starlark.net/starlark"
	"golang.org/x/exp/errors"
)

const (
	ziggyKeyImpl    = "implementation"
	ziggyKeyAttrs   = "attrs?"
	ziggyKeyDeps    = "deps?"
	ziggyKeyOutputs = "outputs?"
	ziggyKeyName    = "name"
	ziggyKeyCtx     = "ctx"
)

type ziggy struct {
	ctx     build.Context
	rt      Runtime
	tasks   map[string]build.Task
	globals starlark.StringDict
}

type Runtime interface {
	build.Runtime

	Sys() starlark.StringDict
}

func New(ctx build.Context, rt Runtime) (build.Store, error) {
	return &ziggy{
		ctx: ctx,
		rt:  rt,
	}, nil
}

func (z *ziggy) GetTask(u *url.URL) (build.Task, error) {
	pkg, err := Load(u, z.ctx)
	if err != nil {
		return nil, errors.New("get target")
	}
	if _, err := pkg.Eval(nil); err != nil {
		return nil, fmt.Errorf("can't find task %q", u.String())
	}
	return pkg.GetTask(u)
}

func (z *ziggy) Run(r io.Reader) error {
	thread := &starlark.Thread{
		Name:  "<ziggy.main>",
		Print: func(_ *starlark.Thread, msg string) { fmt.Println(msg) },
	}

	_, err := starlark.ExecFile(thread, "run.bldy", r, z.rt.Sys())
	if err != nil {
		if evalErr, ok := err.(*starlark.EvalError); ok {
			log.Fatal(evalErr.Backtrace())
		}
		log.Fatal(err)
	}

	return err
}

// Package ziggy is the is the VM for stardust scripts which are
// derivied from starlark scripts
package ziggy

import (
	"bldy.build/bldy/src/build"
	"bldy.build/bldy/src/url"

	"go.starlark.net/starlark"
)

type Package interface {
	Eval(thread *starlark.Thread) (starlark.StringDict, error)
	GetTask(u *url.URL) (build.Task, error)
}

package ziggy

import (
	"bldy.build/bldy/src/build"
	"bldy.build/bldy/src/url"
)

type Task struct {
	name string
	u    url.URL
}

func (t *Task) Name() string {
	return t.name
}

func (t *Task) Dependencies() []*url.URL {
	return nil
}

func (t *Task) Outputs() []string {
	panic("not implemented")
}

func (t *Task) Hash() []byte {
	return []byte{}
}

func (t *Task) Run(build.Runtime) error {
	return nil
}

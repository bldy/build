package skylark

import (
	"fmt"

	"bldy.build/build/executor"
	"github.com/google/skylark"
	"github.com/pkg/errors"
)

// BldyFunc represents a callable SkylarkFunc that will run by bldy
type BldyFunc func(thread *skylark.Thread, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error)

type action struct {
	name           string
	a              BldyFunc
	actionRecorder *actionRecorder
}

func (a *action) Name() string          { return a.name }
func (a *action) Freeze()               {}
func (a *action) Truth() skylark.Bool   { return true }
func (a *action) String() string        { return fmt.Sprintf("action.%s", a.name) }
func (a *action) Type() string          { return fmt.Sprintf("actions.%s", a.name) }
func (a *action) Hash() (uint32, error) { return hashString(a.name), nil }

func (a *action) Call(thread *skylark.Thread, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {

	var i executor.Action
	switch a.name {
	case "run":
		i = &run{}
	case "do_nothing":
		i = &doNothing{}
	}
	if err := unpackStruct(i, kwargs); err != nil {
		return skylark.None, errors.Wrap(err, "action.call")
	}
	a.actionRecorder.Record(i)
	return skylark.None, nil
}

type actionRecorder struct{ calls []executor.Action }

func (ar *actionRecorder) Record(a executor.Action) {
	ar.calls = append(ar.calls, a)
}

// newAction retunrs a new action that will be called during building
func newAction(name string, ac *actionRecorder) *action {
	return &action{
		name:           name,
		actionRecorder: ac,
	}
}

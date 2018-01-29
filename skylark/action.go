package skylark

import (
	"fmt"

	"github.com/google/skylark"
	"github.com/pkg/errors"
)

// Action interface is used for deferred actions that get performed
// during the build stage, unlike rules actions are NOT meant to be executed
// in parralel.
type Action interface {
	Do() error
}

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
func (a *action) Hash() (uint32, error) { return hashString(a.name), nil }
func (a *action) Type() string          { return fmt.Sprintf("actions.%s", a.name) }
func (a *action) Call(thread *skylark.Thread, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var i Action
	switch a.name {
	case "run":
		i = &run{}
	case "do_nothing":
		i = &doNothing{}
	}
	if err := unpackStruct(&i, kwargs); err != nil {
		return skylark.None, errors.Wrap(err, "action.call")
	}
	a.actionRecorder.calls = append(a.actionRecorder.calls, i)
	return skylark.None, nil
}

type actionCall struct {
	name   string
	thread *skylark.Thread
	args   skylark.Tuple
	kwargs []skylark.Tuple
}

type actionRecorder struct{ calls []Action }

// newAction retunrs a new action that will be called during building
func newAction(name string, ac *actionRecorder) *action {
	return &action{
		name:           name,
		actionRecorder: ac,
	}
}

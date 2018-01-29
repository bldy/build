package skylark

import (
	"bytes"
	"fmt"

	"github.com/google/skylark"
)

// newContext returns a new bazel build context.
func newContext(name string, attrs bldyDict) *context {
	ac := new(actionRecorder)
	actionsDict := make(bldyDict)
	for _, actionName := range []string{
		"run", "do_nothing",
	} {
		actionsDict[actionName] = newAction(actionName, ac)
	}

	return &context{
		label:          name,
		buf:            bytes.NewBuffer(nil),
		attrs:          attrs,
		actions:        actionsDict,
		actionRecorder: ac,
	}
}

type context struct {
	label          string
	buf            *bytes.Buffer
	attrs          bldyDict
	actions        bldyDict
	actionRecorder *actionRecorder
}

func (ctx *context) Name() string                             { return "ctx" }
func (ctx *context) Freeze()                                  {}
func (ctx *context) Truth() skylark.Bool                      { return true }
func (ctx *context) String() string                           { return fmt.Sprintf("label = %q", ctx.label) }
func (ctx *context) Hash() (uint32, error)                    { return 0, errDoesntHash }
func (ctx *context) Type() string                             { return "ctx" }
func (ctx *context) AttrNames() []string                      { return nil }
func (ctx *context) Print(thread *skylark.Thread, msg string) { ctx.buf.WriteString(msg) }
func (ctx *context) Attr(name string) (skylark.Value, error) {
	switch name {
	case "label":
		return skylark.String(ctx.label), nil
	case "attrs":
		return ctx.attrs, nil
	case "actions":
		return ctx.actions, nil
	default:
		return nil, fmt.Errorf("ctx doesn't have field or method %q", name)
	}
}

package skylark

import (
	"bytes"
	"fmt"

	"github.com/google/skylark"
	"github.com/google/skylark/skylarkstruct"
)

// newContext returns a new bazel build context.
func newContext(name string, ruleAttrs *skylark.Dict, ruleOutputs *skylark.Dict, kwargs []skylark.Tuple, wd string) (*context, []string, error) {
	ac := new(actionRecorder)
	actionsDict := skylark.StringDict{}
	for _, actionName := range []string{
		"run", "do_nothing",
	} {
		actionsDict[actionName] = newAction(actionName, ac)
	}
	ctx := &context{
		label:          name,
		buf:            bytes.NewBuffer(nil),
		actions:        skylarkstruct.FromStringDict(skylarkstruct.Default, actionsDict),
		actionRecorder: ac,
	}

	if err := processAttrs(ctx, name, ruleAttrs, kwargs, wd); err != nil {
		return nil, []string{}, err
	}

	if outputs, err := processOutputs(ctx, ruleAttrs, ruleOutputs); err != nil {
		return nil, []string{}, err
	} else {

		return ctx, outputs, nil
	}
}

type context struct {
	label   string
	buf     *bytes.Buffer
	attrs   *skylarkstruct.Struct
	outputs skylark.Value

	files          *skylarkstruct.Struct
	actions        *skylarkstruct.Struct
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
	case "files":
		return ctx.files, nil
	case "outputs":
		return ctx.outputs, nil
	default:
		return nil, fmt.Errorf("ctx doesn't have field or method %q", name)
	}
}

package skylark

import (
	"bytes"
	"fmt"

	"bldy.build/build/label"
	"bldy.build/build/workspace"
	"github.com/google/skylark"
	"github.com/google/skylark/skylarkstruct"
)

type skyIO struct {
	files   []string
	outputs []string
}

// newContext returns a new bazel build context.
func newContext(name string, ruleAttrs *skylark.Dict, ruleOutputs *skylark.Dict, kwargs []skylark.Tuple, lbl label.Label, ws workspace.Workspace) (*context, *skyIO, error) {
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
	skyio := &skyIO{}
	var err error
	if err = processAttrs(ctx, name, ruleAttrs, kwargs, ws.PackageDir(lbl)); err != nil {
		return nil, nil, err
	}

	if skyio.files, err = processFiles(ctx, ruleAttrs, kwargs, ws, lbl); err != nil {
		return nil, nil, err
	}
	if skyio.outputs, err = processOutputs(ctx, ruleAttrs, ruleOutputs); err != nil {

		return nil, nil, err
	}
	return ctx, skyio, nil

}

type context struct {
	label string
	buf   *bytes.Buffer

	attrs   skylark.StringDict
	files   skylark.StringDict
	outputs skylark.Value

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
func (ctx *context) Attrs() *skylarkstruct.Struct {
	return skylarkstruct.FromStringDict(skylark.String("attrs"), ctx.attrs)
}
func (ctx *context) Attr(name string) (skylark.Value, error) {
	switch name {
	case "label":
		return skylark.String(ctx.label), nil
	case "attrs":
		return ctx.Attrs(), nil
	case "actions":
		return ctx.actions, nil
	case "files":
		return skylarkstruct.FromStringDict(skylark.String("files"), ctx.files), nil
	case "outputs":
		return ctx.outputs, nil
	default:
		return nil, fmt.Errorf("ctx doesn't have field or method %q", name)
	}
}

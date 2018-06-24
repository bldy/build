package skylark

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"

	"github.com/google/skylark"
	"github.com/google/skylark/skylarkstruct"
)

// newContext returns a new bazel build context.
func newContext(name string, fattrs *skylark.Dict, kwargs []skylark.Tuple, wd string) (*context, error) {
	ac := new(actionRecorder)
	actionsDict := make(bldyDict)
	for _, actionName := range []string{
		"run", "do_nothing",
	} {
		actionsDict[actionName] = newAction(actionName, ac)
	}

	attrs := make(bldyDict)
	attrs["name"] = skylark.String(name) // this is added to all attrs https://github.com/bazelbuild/examples/blob/master/rules/attributes/printer.bzl#L20

	outputs := skylark.StringDict{}
	files := skylark.StringDict{}
	_ = files

	err := WalkDict(fattrs, func(kw skylark.Value, attr Attribute) error { // check the attributes
		arg, ok := findArg(kw, kwargs)

		name := string(kw.(skylark.String))
		if ok { // try finding the kwarg mentioned in the attribute
			attrs[name] = arg
		} else if attr.HasDefault() { // if the attribute has a default and it's not in kwargs
			attrs[name] = attr.GetDefault()
		}

		switch x := attr.(type) {
		case *outputAttr:
			f, err := asFile(attrs[name], wd)
			if err != nil {
				return errors.Wrap(err, "newcontext")
			}
			outputs[name] = f
		case *labelListAttr:
			if x.AllowFiles {
				f, err := asFileList(attrs[name], wd)
				if err != nil {
					return errors.Wrap(err, "newcontext")

				}
				files[name] = f
			}
		default:
			//		log.Printf("%s %T", kw, attr)

		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &context{
		label:          name,
		buf:            bytes.NewBuffer(nil),
		attrs:          attrs,
		files:          skylarkstruct.FromStringDict(skylarkstruct.Default, files),
		outputs:        skylarkstruct.FromStringDict(skylarkstruct.Default, outputs),
		actions:        actionsDict,
		actionRecorder: ac,
	}, nil
}

type context struct {
	label          string
	buf            *bytes.Buffer
	attrs          bldyDict
	outputs        *skylarkstruct.Struct
	files          *skylarkstruct.Struct
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
	case "files":
		return ctx.files, nil
	case "outputs":
		return ctx.outputs, nil
	default:
		return nil, fmt.Errorf("ctx doesn't have field or method %q", name)
	}
}

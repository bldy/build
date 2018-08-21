package skylark

import (
	"github.com/google/skylark"
	"github.com/google/skylark/skylarkstruct"
	"github.com/pkg/errors"
)

// inconsistent behaviour described here
// https://docs.bazel.build/versions/master/skylark/rules.html#files
// https://docs.bazel.build/versions/master/skylark/lib/ctx.html#outputs
func processOutputs(ctx *context, ruleAttrs *skylark.Dict, ruleOutputs *skylark.Dict) ([]string, error) {
	outputs := []string{}
	if ruleOutputs != nil && len(ruleOutputs.Keys()) > 0 {
		outs := skylark.StringDict{}
		for _, tup := range ruleOutputs.Items() {
			if formatted, err := format(tup[1].(skylark.String), ctx.Attrs()); err == nil {
				outputs = append(outputs, formatted)
				if name, ok := skylark.AsString(tup[0]); ok {
					_ = name
					outs[name] = output(formatted)
				}
			}
		}
		ctx.outputs = skylarkstruct.FromStringDict(skylarkstruct.Default, outs)
		return outputs, nil
	}
	return outputs, nil
}

type output string

func (f output) String() string        { return string(f) }
func (f output) Type() string          { return "file" }
func (f output) Freeze()               {}
func (f output) Truth() skylark.Bool   { return true }
func (f output) Hash() (uint32, error) { return hashString(string(f)), nil }

func (f output) Attr(name string) (skylark.Value, error) {
	switch name {
	case "path":
		return skylark.String(string(f)), nil
	default:
		return nil, errors.New("not implemented")
	}
}

func (f output) AttrNames() []string {
	panic("not implemented")
}

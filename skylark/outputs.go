package skylark

import (
	"github.com/google/skylark"
	"github.com/google/skylark/skylarkstruct"
)

// inconsistent behaviour described here
// https://docs.bazel.build/versions/master/skylark/rules.html#files
// https://docs.bazel.build/versions/master/skylark/lib/ctx.html#outputs
func processOutputs(ctx *context, ruleAttrs *skylark.Dict, ruleOutputs *skylark.Dict) ([]string, error) {
	outputs := []string{}
	if ruleOutputs != nil && len(ruleOutputs.Keys()) > 0 {
		outs := skylark.StringDict{}
		for _, tup := range ruleOutputs.Items() {
			if formatted, err := format(tup[1].(skylark.String), ctx.attrs); err == nil {
				outputs = append(outputs, formatted)
				if name, ok := skylark.AsString(tup[0]); ok {
					outs[name], err = asFile(skylark.String(formatted), "")
				}
			}

		}
		ctx.outputs = skylarkstruct.FromStringDict(skylarkstruct.Default, outs)
		return outputs, nil
	}
	return outputs, nil
}

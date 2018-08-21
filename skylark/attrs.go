package skylark

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/google/skylark"
)

func processAttrs(ctx *context, name string, ruleAttrs *skylark.Dict, kwargs []skylark.Tuple, wd string) error {
	ctx.attrs = skylark.StringDict{}
	ctx.attrs["name"] = skylark.String(name) // this is added to all attrs https://github.com/bazelbuild/examples/blob/master/rules/attributes/printer.bzl#L20

	err := WalkDict(ruleAttrs, func(kw skylark.Value, attr Attribute) error { // check the attributes
		arg, ok := findArg(kw, kwargs)
		name := string(kw.(skylark.String))
		if ok {
		} else if attr.HasDefault() { // if the attribute has a default and it's not in kwargs
			arg = attr.GetDefault()
		} else if attr, ok := attr.(CanAllowEmpty); ok && attr.AllowsEmpty() {
			ctx.attrs[name] = attr.Empty()
			return nil
		} else {
			return fmt.Errorf("attribute %q isn't allowed to be empty", name)
		}

		if converter, ok := attr.(Converts); ok {
			var err error
			ctx.attrs[name], err = converter.Convert(arg)
			if err != nil {
				return err
			}
		} else {
			ctx.attrs[name] = arg
		}
		return nil
	})
	return err
}

// WalkAttrs traverses attributes
func WalkDict(x *skylark.Dict, wf WalkAttrFunc) error {
	if x == nil {
		panic("can't be nil")
	}
	i := x.Iterate()
	var p skylark.Value
	for i.Next(&p) {
		val, _, err := x.Get(p)
		if err != nil {
			l.Println(errors.Wrap(err, "skylarkRule walk attrs"))
		}
		if Attr, ok := val.(Attribute); ok {
			if err := wf(p, Attr); err != nil {
				return err
			}
		}
	}
	return nil
}

type WalkAttrFunc func(skylark.Value, Attribute) error
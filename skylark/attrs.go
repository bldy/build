package skylark

import (
	"fmt"

	"bldy.build/build/label"
	"github.com/pkg/errors"

	"github.com/google/skylark"
	"github.com/google/skylark/skylarkstruct"
)

func processAttrs(ctx *context, name string, ruleAttrs *skylark.Dict, kwargs []skylark.Tuple, wd string) error {
	attrs := skylark.StringDict{}
	attrs["name"] = skylark.String(name) // this is added to all attrs https://github.com/bazelbuild/examples/blob/master/rules/attributes/printer.bzl#L20

	files := skylark.StringDict{}
	_ = files
	err := WalkDict(ruleAttrs, func(kw skylark.Value, attr Attribute) error { // check the attributes
		arg, ok := findArg(kw, kwargs)
		name := string(kw.(skylark.String))
		if ok { // try finding the kwarg mentioned in the attribute
			attrs[name] = arg
		} else if attr.HasDefault() { // if the attribute has a default and it's not in kwargs
			attrs[name] = attr.GetDefault()
		} else if attr, ok := attr.(AllowsEmpty); ok {
			attrs[name] = attr.Empty()
		}

		switch x := attr.(type) {
		case *labelListAttr:
			var err error
			if attrs[name], err = processLabelListAttr(attrs[name]); err != nil {
				return errors.Wrap(err, "processAttrs")
			}
			if x.AllowFiles {
				f, err := asFileList(arg, wd)
				if err != nil {
					return errors.Wrap(err, "newcontext")
				}
				files[name] = f
			}
		}
		return nil
	})
	ctx.files = skylarkstruct.FromStringDict(skylarkstruct.Default, files)
	ctx.attrs = skylarkstruct.FromStringDict(skylarkstruct.Default, attrs)
	return err

}

func processLabelListAttr(v skylark.Value) (*skylark.List, error) {
	if list, ok := v.(*skylark.List); ok {
		lbls := skylark.NewList([]skylark.Value{})
		var p skylark.Value
		i := list.Iterate()
		for i.Next(&p) {
			if s, ok := skylark.AsString(p); ok {
				if lbl, err := label.Parse(s); err == nil {
					lbls.Append(*lbl)

				} else {
					return nil, err
				}
			}
		}
		i.Done()
		return lbls, nil
	}
	return nil, fmt.Errorf("label list attribute only accept lists not %T", v)
}
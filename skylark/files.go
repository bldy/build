package skylark

import (
	"fmt"

	"bldy.build/build/file"
	"bldy.build/build/label"
	"bldy.build/build/workspace"
	"github.com/google/skylark"
)

// this is very busy
func processFiles(ctx *context, ruleAttrs *skylark.Dict, kwargs []skylark.Tuple, ws workspace.Workspace, lbl label.Label) ([]string, error) {
	files := []string{}
	ctx.files = skylark.StringDict{}
	err := WalkDict(ruleAttrs, func(kw skylark.Value, attr Attribute) error { // check the attributes
		name := string(kw.(skylark.String))
		arg := ctx.attrs[name]

		switch x := attr.(type) {
		case *labelAttr:
			l, ok := arg.(label.Label)
			if !ok {
				return fmt.Errorf("attribute %q should be of type list consisting of strings", name)
			}
			if x.AllowFiles {
				f := file.New(l, lbl, ws)
				if f.Exists() {
					files = append(files, f.Path())
				}
				ctx.files[name] = f
			}
		case *labelListAttr:
			if x.AllowFiles {
				lblList, ok := arg.(*skylark.List)
				if !ok {
					return fmt.Errorf("attribute %q should be of type list consisting of strings", name)
				}
				skyfiles := skylark.NewList([]skylark.Value{})

				i := lblList.Iterate()
				var p skylark.Value
				for i.Next(&p) {
					l, ok := p.(label.Label)
					if !ok {
						return fmt.Errorf("label list only allows strings not %Ts: %v", p.Type(), p)
					}
					f := file.New(l, lbl, ws)
					if f.Exists() {
						files = append(files, f.Path())
					}

					skyfiles.Append(f)
				}
				ctx.files[name] = skyfiles
			}
		}
		return nil
	})
	return files, err
}
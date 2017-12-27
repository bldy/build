package skylark

import (
	"fmt"

	"github.com/google/skylark"
)

type skylarkContext struct {
	label string
	attrs *skylark.Dict
}

func (ctx *skylarkContext) Name() string          { return "ctx" }
func (ctx *skylarkContext) Freeze()               {}
func (ctx *skylarkContext) Truth() skylark.Bool   { return true }
func (ctx *skylarkContext) String() string        { return fmt.Sprintf("label = %q", ctx.label) }
func (ctx *skylarkContext) Hash() (uint32, error) { return 12312, nil }
func (ctx *skylarkContext) Type() string          { return "ctx" }
func (ctx *skylarkContext) AttrNames() []string   { return nil }
func (ctx *skylarkContext) Attr(name string) (skylark.Value, error) {
	switch name {
	case "label":
		return skylark.String(ctx.label), nil
	case "attrs":
		return ctx.attrs, nil
	default:
		return nil, fmt.Errorf("ctx doesn't have field or method %q", name)
	}
} // returns (nil, nil) if attribute not present

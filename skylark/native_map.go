package skylark

import (
	"fmt"

	"github.com/google/skylark"
	"github.com/pkg/errors"
)

type nativeMap skylark.StringDict

func (x nativeMap) Freeze()               {}
func (x nativeMap) Hash() (uint32, error) { return 0, errors.New("ctx does not implement hash") }
func (x nativeMap) String() string        { return "" }
func (x nativeMap) Truth() skylark.Bool   { return true }
func (x nativeMap) Type() string          { return "attrs" }
func (x nativeMap) AttrNames() []string   { return nil }
func (x nativeMap) Attr(name string) (skylark.Value, error) {
	if v, ok := x[name]; ok {
		return v, nil
	}
	return skylark.None, fmt.Errorf("ctx.attrs has no %s field or method", name)
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

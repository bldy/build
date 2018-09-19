package skylark

import (
	"errors"
	"fmt"

	"github.com/google/skylark"
)

type bldyDict skylark.StringDict

func (x bldyDict) Freeze()               {}
func (x bldyDict) Hash() (uint32, error) { return 0, errors.New("ctx does not implement hash") }
func (x bldyDict) String() string        { return "" }
func (x bldyDict) Truth() skylark.Bool   { return true }
func (x bldyDict) Type() string          { return "bldydict" }
func (x bldyDict) AttrNames() []string   { return nil }
func (x bldyDict) Attr(name string) (skylark.Value, error) {
	if v, ok := x[name]; ok {
		return v, nil
	}
	return skylark.None, fmt.Errorf("ctx.attrs has no %s field or method", name)
}

// NewBldyDict returns a new skylark.StringDict that can be used as a skylark.Value
func NewBldyDict() bldyDict {
	return make(bldyDict)
}

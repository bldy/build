package skylark

import (
	"fmt"

	"github.com/google/skylark"
)

type attributer struct{}

var attrFuncs = map[string]*skylark.Builtin{
	"int":        skylark.NewBuiltin("int", attrInt),
	"label_list": skylark.NewBuiltin("label_list", attrLabelList),
}

func (a attributer) Name() string          { return "attributer" }
func (a attributer) Hash() (uint32, error) { return 0, nil }
func (a attributer) Freeze()               {}
func (a attributer) String() string        { return "" }
func (a attributer) Type() string          { return "attributer" }
func (a attributer) Truth() skylark.Bool   { return true }

func (a attributer) Attr(name string) (skylark.Value, error) {
	if f, ok := attrFuncs[name]; ok {
		return f, nil
	}
	return skylark.None, fmt.Errorf("attr doen't implement %s", name)
}
func (a attributer) AttrNames() (names []string) {
	for k := range attrFuncs {
		names = append(names, k)
	}
	return
}

func attrLabelList(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	allowFiles := false
	err := skylark.UnpackArgs(fn.Name(), args, kwargs, "allow_files", &allowFiles)
	if err != nil {
		return nil, err
	}
	return &labelAttr{allowFiles: allowFiles}, nil
}

func attrInt(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var d int
	err := skylark.UnpackArgs(fn.Name(), args, kwargs, "default", &d)
	if err != nil {
		return nil, err
	}

	return &intAttr{def: d}, nil
}

type attr interface {
	Apply(skylark.Value) error
}
type intAttr struct {
	def int
}

func (a *intAttr) Name() string          { return "int_attr" }
func (a *intAttr) Hash() (uint32, error) { return 0, nil }
func (a *intAttr) Freeze()               {}
func (a *intAttr) String() string        { return fmt.Sprintf("%s.default = %v", a.Name(), a.def) }
func (a *intAttr) Type() string          { return "int_attr" }
func (a *intAttr) Truth() skylark.Bool   { return true }

type labelAttr struct {
	allowFiles bool
}

func (a *labelAttr) Name() string          { return "label_attr" }
func (a *labelAttr) Hash() (uint32, error) { return 0, nil }
func (a *labelAttr) Freeze()               {}
func (a *labelAttr) String() string        { return fmt.Sprintf("%s.allow_files = %v", a.Name(), a.allowFiles) }
func (a *labelAttr) Type() string          { return "label_attr" }
func (a *labelAttr) Truth() skylark.Bool   { return true }

package skylark

import (
	"fmt"

	"github.com/google/skylark"
)

type attrType int

const (
	_int attrType = iota
	_labelList
)

type attr struct {
	t   attrType
	def skylark.Value
}

func (a *attr) Name() string          { return "attr" }
func (a *attr) Hash() (uint32, error) { return 0, nil }
func (a *attr) Freeze()               {}
func (a *attr) String() string        { return fmt.Sprintf("%s.type = %v", a.Name(), a.t.String()[1:]) }
func (a *attr) Type() string          { return "int_attr" }
func (a *attr) Truth() skylark.Bool   { return true }

func attrLabelList(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	allowFiles := false
	err := skylark.UnpackArgs(fn.Name(), args, kwargs, "allow_files", &allowFiles)
	if err != nil {
		return nil, err
	}
	return &attr{t: _labelList}, nil
}

func attrInt(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var d int
	err := skylark.UnpackArgs(fn.Name(), args, kwargs, "default", &d)
	if err != nil {
		return nil, err
	}
	return &attr{t: _int, def: skylark.MakeInt(d)}, nil
}

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

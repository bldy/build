package skylark

import (
	"fmt"
	"log"
	"os"

	"github.com/google/skylark"
	"github.com/pkg/errors"
)

type attrType int

const (
	_int attrType = iota
	_labelList
)

type attr struct {
	t          attrType
	def        skylark.Value
	allowFiles bool
}

func (a *attr) Name() string          { return "attr" }
func (a *attr) Hash() (uint32, error) { return 0, errors.New("attr doesn't implenent hash") }
func (a *attr) Freeze()               {}
func (a *attr) String() string {
	return fmt.Sprintf("attr.type = %v, attr.has_default = %v", a.t.String()[1:], a.def != nil)
}
func (a *attr) Type() string        { return "attr" }
func (a *attr) Truth() skylark.Bool { return true }

func newAttr(t attrType, thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	attr := &attr{t: t}
	err := skylark.UnpackArgs(fn.Name(), args, kwargs, "allow_files", &attr.allowFiles, "default", &attr.def)
	if err != nil && os.Getenv("BLDY_DEBUG") == "1" {
		log.Println(err)
	}
	return attr, nil
}

func attrInt(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	return newAttr(_int, thread, fn, args, kwargs)
}
func attrLabelList(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	return newAttr(_labelList, thread, fn, args, kwargs)
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

package skylark

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/pkg/errors"

	"github.com/google/skylark"
)

func asFileList(f skylark.Value, wd string) (*skylark.List, error) {
	var vals []skylark.Value
	switch x := f.(type) {
	case *skylark.List:
		var p skylark.Value
		i := x.Iterate()
		for i.Next(&p) {
			f, err := asFile(p, wd)
			if err != nil {
				return nil, errors.Wrap(err, "asfilelist")
			}
			vals = append(vals, f)
		}
	default:
		return nil, fmt.Errorf("can't convert %T to file", f)
	}
	return skylark.NewList(vals), nil
}

func asFile(f skylark.Value, wd string) (*File, error) {
	if name, ok := skylark.AsString(f); ok {
		return &File{name: name, wd: wd}, nil
	}

	return nil, fmt.Errorf("can't convert %T to file", f)
}

// File is a bazel file
// https://docs.bazel.build/versions/master/skylark/lib/File.html#modules.File
type File struct {
	name string
	wd   string
}

func (f *File) Path() string {
	return path.Join(f.wd, f.name)
}
func (f *File) Name() string        { return f.name }
func (f *File) Freeze()             {}
func (f *File) Truth() skylark.Bool { return true }
func (f *File) String() string      { return f.name }
func (f *File) Type() string        { return "File" }
func (f *File) Hash() (uint32, error) {
	bytz, err := ioutil.ReadFile(f.Path())
	if err != nil {
		return 0, err
	}
	return hashString(string(bytz)), errDoesntHash
}
func (f *File) AttrNames() []string { return nil }
func (f *File) Attr(name string) (skylark.Value, error) {
	switch name {
	case "path":
		return skylark.String(f.Path()), nil
	default:
		return nil, fmt.Errorf("ctx doesn't have field or method %q", name)
	}
}

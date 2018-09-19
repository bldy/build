package file

import (
	"fmt"
	"os"
	"path/filepath"

	"bldy.build/build/label"
	"bldy.build/build/workspace"
	"github.com/google/skylark"
)

// File is a bazel file
// https://docs.bazel.build/versions/master/skylark/lib/File.html#modules.File
type File struct {
	isSource bool
	pkg      label.Label
	file     label.Label
	ws       workspace.Workspace
}

func New(file, pkg label.Label, ws workspace.Workspace) *File {
	return &File{
		file: file,
		pkg:  pkg,
		ws:   ws,
	}
}

func (f *File) Exists() bool {
	_, err := os.Stat(f.Path())
	return err == nil
}

func (f *File) Path() string {

	if f.file.IsAbs() {
		return f.ws.File(f.file)
	} else {
		if err := f.file.Valid(); err != nil {
			panic(err)
		}
		return filepath.Join(f.ws.PackageDir(f.pkg), f.file.Package(), f.file.Name())
	}
}
func (f *File) Name() string        { return f.file.Name() }
func (f *File) Freeze()             {}
func (f *File) Truth() skylark.Bool { return true }
func (f *File) String() string      { return f.file.Name() }
func (f *File) Type() string        { return "File" }
func (f *File) Hash() (uint32, error) {
	return hashString(f.Path()), nil
}
func (f *File) AttrNames() []string { return nil }
func (f *File) Attr(name string) (skylark.Value, error) {
	switch name {
	case "is_source":
		_, err := os.Stat(f.Path())
		return !skylark.Bool(os.IsNotExist(err)), nil
	case "name":
		return skylark.String(f.file), nil
	case "path":
		return skylark.String(f.Path()), nil
	default:
		return nil, fmt.Errorf("ctx doesn't have field or method %q", name)
	}
}

// hashString computes the FNV hash of s.
// copied for consistency https://github.com/google/skylark/blob/f09c8ae6985f50d0fbdd2a30e4c3cfff3c4746ce/hashtable.go#L335
func hashString(s string) uint32 {
	var h uint32
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

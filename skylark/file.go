package skylark

import (
	"io/ioutil"

	"github.com/google/skylark"
)

// File is a bazel file
// https://docs.bazel.build/versions/master/skylark/lib/File.html#modules.File
type File struct {
	file string
}

func (f *File) Name() string        { return f.file }
func (f *File) Freeze()             {}
func (f *File) Truth() skylark.Bool { return true }
func (f *File) String() string      { return f.file }
func (f *File) Type() string        { return "File" }
func (f *File) Hash() (uint32, error) {
	bytz, err := ioutil.ReadFile(f.file)
	if err != nil {
		return 0, err
	}
	return hashString(string(bytz)), errDoesntHash
}

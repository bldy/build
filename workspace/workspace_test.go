package workspace

import (
	"go/build"
	"path"
	"testing"
)

var tests = []struct {
	name string
	dir  string
	wd   string
	err  error
}{
	{
		name: "testdata",
		dir:  path.Join(build.Default.GOPATH, "src", "bldy.build", "build", "tests", "testdata"),
		err:  nil,
		wd:   path.Join(build.Default.GOPATH, "src", "bldy.build", "build", "tests", "testdata"),
	},
	{
		name: "testdatafromempty",
		dir:  path.Join(build.Default.GOPATH, "src", "bldy.build", "build", "tests", "testdata", "empty"),
		err:  nil,
		wd:   path.Join(build.Default.GOPATH, "src", "bldy.build", "build", "tests", "testdata"),
	},
	{
		name: "temp",
		dir:  "/tmp/",
		err:  ErrNotAWorkspace,
		wd:   "",
	},
}

func TestNew(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ws, err := New(test.dir)
			if expected, got := test.err, err; expected != got {
				t.Logf(`was expecting "%v" got "%v" instead`, expected, got)
				t.Fail()
			}
			if err != nil {
				return
			}
			if expected, got := test.wd, ws.AbsPath(); expected != got {
				t.Logf("was expecting %q got %q instead", expected, got)
			}
		})
	}
}

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
	err  bool
}{
	{
		name: "testdata",
		dir:  path.Join(build.Default.GOPATH, "src", "bldy.build", "build", "integration", "testdata"),
		wd:   path.Join(build.Default.GOPATH, "src", "bldy.build", "build", "integration", "testdata"),
	},
	{
		name: "testdatafromempty",
		dir:  path.Join(build.Default.GOPATH, "src", "bldy.build", "build", "integration", "testdata", "empty"),
		wd:   path.Join(build.Default.GOPATH, "src", "bldy.build", "build", "integration", "testdata"),
	},
	{
		name: "temp",
		dir:  "/tmp/",
		err:  true,
		wd:   "",
	},
}

func TestNew(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ws, err := New(test.dir)
			if (err == nil) == test.err {
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

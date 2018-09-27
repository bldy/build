package workspace

import (
	"os"
	"path"
	"testing"
)

func TestNew(t *testing.T) {
	wd, _ := os.Getwd()
	var tests = []struct {
		name string
		dir  string
		wd   string
		err  bool
	}{
		{
			name: "testdata",
			dir:  path.Join(wd, "..", "tests", "testdata"),
			wd:   path.Join(wd, "..", "tests", "testdata"),
		},
		{
			name: "testdatafromempty",
			dir:  path.Join(wd, "..", "tests", "testdata", "empty"),
			wd:   path.Join(wd, "..", "tests", "testdata"),
		},
		{
			name: "temp",
			dir:  "/tmp/",
			err:  true,
			wd:   "",
		},
	}
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

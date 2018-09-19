package file

import (
	"testing"

	"bldy.build/build/workspace/testws"

	"bldy.build/build/label"
)

func TestPath(t *testing.T) {
	tests := []struct {
		name string
		file string
		pkg  string
		wd   string
		path string
	}{
		{
			name: "fileAtRoot",
			file: "a.c",
			pkg:  "//:",
			wd:   "/home/x/src/awesomeproject/",
			path: "/home/x/src/awesomeproject/a.c",
		},
		{
			name: "fileAbsolute",
			file: "//:a.c",
			pkg:  "//:",
			wd:   "/home/x/src/awesomeproject/",
			path: "/home/x/src/awesomeproject/a.c",
		},
		{
			name: "fileInPkg",
			file: "a.c",
			pkg:  "//b:",
			wd:   "/home/x/src/awesomeproject/",
			path: "/home/x/src/awesomeproject/b/a.c",
		},
		{
			name: "fileInDir",
			file: "b/a.c",
			pkg:  "//.:",
			wd:   "/home/x/src/awesomeproject/",
			path: "/home/x/src/awesomeproject/b/a.c",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, _ := label.Parse(test.file) // these aren't the errors we are testing for
			pkg, _ := label.Parse(test.pkg)   //
			f := New(file, pkg, &testws.TestWS{WD: test.wd})
			if expected, got := test.path, f.Path(); expected != got {
				t.Logf("was expecting %q got %q instead", expected, got)
				t.Fail()
			}
		})
	}
}
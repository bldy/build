package skylark

import (
	"os"
	"path"
	"testing"

	"bldy.build/build/label"
	"bldy.build/build/skylark/skylarkutils"
	"bldy.build/build/workspace"

	"github.com/google/skylark"
)

func TestEvalGlob(t *testing.T) {
	wd, _ := os.Getwd()
	tests := []struct {
		name  string
		wd    string
		label string
		err   error
		files []string
	}{
		{
			name:  "noexclusions",
			wd:    "glob",
			label: "//.:glob",
			files: []string{
				"1.xml",
				"2.xml",
				"3.xml",
			},
			err: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wd = path.Join(wd, "testdata", test.wd)
			ws, err := workspace.New(wd)

			if err != nil {
				t.Log(err)
				t.Fail()
				return
			}
			vm, _ := New(ws)
			l, _ := label.Parse(test.label)

			target, err := vm.GetTarget(l)
			if test.err != errAny && err != test.err {
				t.Log(err)
				t.Fail()
				return
			}
			if test.err != errAny && target == nil {
				t.Fail()
				return
			}
			r, ok := target.(*Rule)
			if !ok {
				t.Fail()
				return

			}

			n, ok := findArg(skylark.String("srcs"), r.KWArgs)
			if !ok {
				t.Log("no sources")
				t.Fail()
				return
			}
			srcs, ok := n.(*skylark.List)
			if !ok {
				t.Log("sources is not a list")
				t.Fail()
				return
			}
			//		for i, f := range test.files {
			//			test.files[i] = path.Join(wd, f)
			//		}
			if ok, err := compareLists(srcs, test.files); err != nil {
				t.Log(err)
				t.Fail()
				return
			} else if !ok {
				t.Log("glob output and expected list of files are not the same")

				t.Fail()
				return
			}

		})
	}
}

func compareLists(a *skylark.List, b []string) (bool, error) {
	i, err := skylarkutils.ListToGo(a)
	if err != nil {
		return false, err
	}
	if list, ok := i.([]string); ok {
		if len(list) != len(b) {
			return false, nil
		}
		for i, s := range list {
			if s != b[i] {
				return false, nil
			}
		}
	}
	return true, nil
}
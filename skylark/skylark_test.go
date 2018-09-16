package skylark

import (
	"errors"
	"os"
	"path"
	"testing"

	"bldy.build/build/label"
	"bldy.build/build/workspace"
)

var errAny = errors.New("any error")

func TestEval(t *testing.T) {
	tests := []struct {
		name  string
		wd    string
		label string
		err   error
	}{
		{
			name:  "notexist",
			wd:    "",
			label: "//notexist:libsncmds",
			err:   errAny,
		},
		{
			name:  "empty",
			wd:    "empty",
			label: "//.:test",
			err:   nil,
		},
		{
			name:  "noop_skylark",
			wd:    "noop",
			label: "//.:noop",
			err:   nil,
		},
		{
			name:  "printer",
			wd:    "printer",
			label: "//.:print",
			err:   nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wd, _ := os.Getwd()
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
			if target != nil {
			} else if test.err == nil && target == nil {
				t.Fail()
				return
			}
		})
	}
}

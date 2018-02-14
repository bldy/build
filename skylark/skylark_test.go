package skylark

import (
	"errors"
	"os"
	"testing"

	"bldy.build/build/label"
	"bldy.build/build/workspace"
)

var errAny = errors.New("any error")

func TestEval(t *testing.T) {
	tests := []struct {
		name  string
		label string
		err   error
	}{
		{
			name:  "notexist",
			label: "//notexist:libsncmds",
			err:   errAny,
		},
		{
			name:  "simple_skylark",
			label: "//testdata:test",
			err:   nil,
		},
		{
			name:  "noop_skylark",
			label: "//testdata:noop",
			err:   nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wd, _ := os.Getwd()
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
				//		log.Println(prettyprint.AsJSON(target))
			}
		})
	}
}

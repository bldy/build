package skylark

import (
	"errors"
	"os"
	"testing"

	"bldy.build/build/label"
	_ "bldy.build/build/rules/cc"
)

func TestNew(t *testing.T) {
	wd, _ := os.Getwd()
	_, err := New(wd)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

var errAny = errors.New("any error")

func TestEval(t *testing.T) {
	tests := []struct {
		name  string
		label string
		err   error
	}{
		{
			name:  "simple",
			label: "//skylark/testdata:test",
			err:   nil,
		},
		{
			name:  "some_target",
			label: "//skylark/testdata:some_target",
			err:   nil,
		},
		{
			name:  "notexist",
			label: "//skylark/notexist:libsncmds",
			err:   errAny,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wd, _ := os.Getwd()
			vm, _ := New(wd + "skylark")
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
		})
	}
}

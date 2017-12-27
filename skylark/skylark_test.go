package skylark

import (
	"errors"
	"log"
	"os"
	"testing"

	"bldy.build/build/label"
	_ "bldy.build/build/rules/cc"
	"sevki.org/lib/prettyprint"
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
			name:  "notexist",
			label: "//skylark/notexist:libsncmds",
			err:   errAny,
		},
		{
			name:  "simple_skylark",
			label: "//skylark/testdata:test",
			err:   nil,
		},
		{
			name:  "some_target",
			label: "//skylark/testdata:some_target",
			err:   nil,
		},
		{
			name:  "native_cc",
			label: "//skylark/testdata:native_cc",
			err:   nil,
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
			if target != nil {
				log.Println(prettyprint.AsJSON(target))
			}
		})
	}
}

package skylark

import (
	"errors"
	"os"
	"testing"

	_ "bldy.build/build/rules/cc"
	"bldy.build/build/url"
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
		name string
		url  string
		err  error
	}{
		{
			name: "simple",
			url:  "//skylark/testdata:test",
			err:  nil,
		},
		{
			name: "some_target",
			url:  "//skylark/testdata:some_target",
			err:  nil,
		},
		{
			name: "notexist",
			url:  "//skylark/notexist:libsncmds",
			err:  errAny,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wd, _ := os.Getwd()
			vm, _ := New(wd + "skylark")
			u, _ := url.Parse(test.url)
			target, err := vm.GetTarget(u)
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

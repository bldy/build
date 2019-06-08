package ziggy

import (
	"os"
	"path"
	"testing"

	"bldy.build/bldy/src/build"
	"bldy.build/bldy/src/url"
)

func TestDoesntExist(t *testing.T) {
	u, err := url.Parse("empty#empty")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	z := New("", build.DefaultContext)
	r, err := z.GetTarget(u)
	if err == nil || r != nil {
		t.Log("did not expwct a target")
		t.Fail()
	}
}

func TestEval(t *testing.T) {
	wd, _ := os.Getwd()
	tests := []struct {
		name string
		url  string
		wd   string
		err  error
	}{
		{
			name: "empty",
			url:  "empty",
			wd:   path.Join(wd, "testdata"),
			err:  nil,
		},
		{
			name: "context_tester",
			url:  "ctx#context_tester",
			wd:   path.Join(wd, "testdata"),
			err:  nil,
		},
		{
			name: "run",
			url:  "run#run_test",
			wd:   path.Join(wd, "testdata"),
			err:  nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := New(test.wd, build.DefaultContext)
			u, _ := url.Parse(test.url)
			target, err := vm.GetTarget(u)
			if err != test.err {
				t.Log(err)
				t.Fail()
				return
			}
			if target == nil {
				t.Fail()
				return
			}
			if target == nil {
				t.Fail()
				return
			}
		})
	}
}
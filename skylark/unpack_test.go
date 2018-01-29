package skylark

import (
	"reflect"
	"testing"

	"github.com/google/skylark"
	"github.com/kr/pretty"
)

// makes tests more compact
type str = skylark.String
type tup = skylark.Tuple

func eqStringSlice(ai, bi interface{}) bool {
	a := reflect.ValueOf(ai)
	b := reflect.ValueOf(bi)
	if !a.IsValid() && !b.IsValid() {
		return true
	}
	if a.Len() != b.Len() {
		return false
	}

	for i := 0; i < a.Len(); i++ {
		if a.Index(i) != b.Index(i) {
			return false
		}
	}

	return true
}

func TestListToGo(t *testing.T) {
	tests := []struct {
		name string
		l    *skylark.List
		i    interface{}
	}{
		{
			name: "string",
			l: skylark.NewList(tup{
				str("x.txt"), str("y.txt"),
			}),
			i: []string{"x.txt", "y.txt"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, _ := skylarkListToGo(test.l)
			if eqStringSlice(test.i, s) {
				t.Logf("test != i\n%s", pretty.Diff(test.i, s))
				t.Fail()
			}
		})
	}
}

func TestUnpack(t *testing.T) {
	tests := []struct {
		name   string
		kwargs []skylark.Tuple
		i      interface{}
	}{
		{
			name: "string",
			kwargs: []tup{
				tup{
					str("executable"),
					str("gcc"),
				},
			},
			i: &run{Executable: "gcc"},
		},
		{
			name: "bool",
			kwargs: []tup{
				tup{
					str("use_default_shell_env"),
					skylark.Bool(true),
				},
			},
			i: &run{UseDefaultShellEnv: true},
		},
		{
			name: "list",
			kwargs: []tup{
				tup{
					str("files"),
					skylark.NewList([]skylark.Value{
						str("x.txt"), str("y.txt"),
					}),
				},
			},
			i: &run{Files: []string{"x.txt", "y.txt"}},
		},
	}
	for _, test := range tests {
		typ := reflect.TypeOf(test.i).Elem()
		t.Run(test.name, func(t *testing.T) {
			i := reflect.New(typ).Interface()

			err := unpackStruct(i, test.kwargs)
			if err != nil {
				t.Log(err)
			}
			equal := reflect.DeepEqual(test.i, i)
			if !equal {
				t.Logf("test != i\n%s", pretty.Diff(test.i, i))
				t.Fail()
			}
		})
	}
}

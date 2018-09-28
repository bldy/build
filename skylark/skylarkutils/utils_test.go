
package skylarkutils

import (
	"reflect"
	"testing"

	"github.com/google/skylark"
	"github.com/kr/pretty"
)

// makes tests more compact
type str = skylark.String
type tup = skylark.Tuple

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
			s, _ := ListToGo(test.l)
			if eqStringSlice(test.i, s) {
				t.Logf("test != i\n%s", pretty.Diff(test.i, s))
				t.Fail()
			}
		})
	}
}

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
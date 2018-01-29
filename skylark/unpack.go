package skylark

import (
	"errors"
	"fmt"
	"reflect"

	"bitbucket.org/pkg/inflect"
	"github.com/google/skylark"
)

func skylarkListToGo(x *skylark.List) (interface{}, error) {
	var vals interface{}
	var p skylark.Value
	it := x.Iterate()
	for it.Next(&p) {
		v, err := skylarkToGo(p)
		if err != nil {
			return err, nil
		}
		switch n := v.(type) {
		case string:
			if vals == nil {
				vals = []string{}
			}
			vals = append(vals.([]string), n)
		}
	}
	return vals, nil
}

func skylarkToGo(i interface{}) (interface{}, error) {
	switch x := i.(type) {
	case skylark.String:
		return string(x), nil
	case skylark.Bool:
		return bool(x), nil
	case *skylark.List:
		return skylarkListToGo(x)
	case skylark.Int:
		if n, ok := x.Int64(); ok {
			return n, nil
		}
		if n, ok := x.Uint64(); ok {
			return n, nil
		}
		return 0, nil
	default:
		return nil, errors.New("can't convert skylark value to go value")
	}
}

// unpackStruct takes kwargs in the form of []skylark.Tuples
// and unpacks its values in to a struct.
//
// There are some caveats in this process that is the result of
// limitations in Go.
//
// Since Go's reflect package doesn't allow setting values of unexported fields
// this function will attempt to use the inflect.Typeify function to convert
// python style identifiers to go style.
func unpackStruct(i interface{}, kwargs []skylark.Tuple) error {
	v := reflect.ValueOf(i).Elem()
	for _, kwarg := range kwargs {
		name := string(kwarg.Index(0).(skylark.String)) // first is the name
		value := kwarg.Index(1)

		// TODO(sevki): add tags to make it more go-ey.
		field := v.FieldByName(inflect.Camelize(name))
		if !field.IsValid() {
			return fmt.Errorf("%T doesn't have a field called %s", i, inflect.Camelize(name))
		}

		val, err := skylarkToGo(value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(val))
	}
	return nil
}

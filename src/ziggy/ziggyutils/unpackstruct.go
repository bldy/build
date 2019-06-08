package ziggyutils

import (
	"errors"
	"fmt"
	"reflect"

	"go.starlark.net/starlark"
)

var starvalue = reflect.TypeOf((*starlark.Value)(nil)).Elem()

// unpackStruct takes kwargs in the form of []skylark.Tuples
// and unpacks its values in to a struct.
//
// There are some caveats in this process that is the result of
// limitations in Go.
//
// Since Go's reflect package doesn't allow setting values of unexported fields
// this function will attempt to use the inflect.Typeify function to convert
// python style identifiers to go style.
func UnpackStruct(i interface{}, kwargs []starlark.Tuple) error {

	t := reflect.TypeOf(i).Elem()

	fieldsByTag := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("ziggy")

		fieldsByTag[tag] = field.Name
	}
	v := reflect.ValueOf(i).Elem()
	for _, kwarg := range kwargs {
		name := string(kwarg.Index(0).(starlark.String)) // first is the name
		value := kwarg.Index(1)

		field := v.FieldByName(fieldsByTag[name])

		if !field.IsValid() {
			return fmt.Errorf("%T doesn't have a field called %s", i, name)
		}
		var val interface{}

		if !field.Type().Implements(starvalue) {
			var err error
			val, err = ValueToGo(value)
			if err != nil {
				return err
			}
		} else {
			val = value
		}
		field.Set(reflect.ValueOf(val))
	}

	return nil
	return nil

}

func ListToGo(x *starlark.List) (interface{}, error) {
	if x == nil {
		return nil, errors.New("list does not exist")
	}
	var vals interface{}
	var p starlark.Value
	it := x.Iterate()
	for it.Next(&p) {
		v, err := ValueToGo(p)
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

func ValueToGo(i interface{}) (interface{}, error) {
	switch x := i.(type) {
	case starlark.String:
		return string(x), nil
	case starlark.Bool:
		return bool(x), nil
	case *starlark.List:
		return ListToGo(x)
	case starlark.Int:
		if n, ok := x.Int64(); ok {
			return n, nil
		}
		if n, ok := x.Uint64(); ok {
			return n, nil
		}
		return 0, nil
	default:
		return nil, fmt.Errorf("can't convert skylark value %T to go value", i)
	}
}
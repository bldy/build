package skylarkutils

import (
	"errors"
	"fmt"

	"bldy.build/build/label"

	"bldy.build/build/file"
	"github.com/google/skylark"
)

func ListToGo(x *skylark.List) (interface{}, error) {
	if x == nil {
		return nil, errors.New("list does not exist")
	}
	var vals interface{}
	var p skylark.Value
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
	case label.Label:
		return string(x), nil
	case skylark.String:
		return string(x), nil
	case skylark.Bool:
		return bool(x), nil
	case *skylark.List:
		return ListToGo(x)
	case *file.File:
		return x.Path(), nil
	case skylark.Int:
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
package skylarkutils

import (
	"fmt"

	"github.com/pkg/errors"

	"bldy.build/build/file"
	"bldy.build/build/label"
	"github.com/google/skylark"
)

func DictToStringDict(x *skylark.Dict) (skylark.StringDict, error) {
	if x == nil {
		return nil, errors.New("map does not exist")
	}
	vals := make(skylark.StringDict)
	for _, k := range x.Keys() {
		key, ok := skylark.AsString(k)
		if !ok {
			return nil, errors.New("key must be a string")
		}
		v, ok, err := x.Get(k)
		if err != nil {
			return nil, fmt.Errorf("couldn't find value a value for %q in the dict", key)
		}
		vals[key] = v
	}
	return vals, nil
}

func DictToMap(x *skylark.Dict) (map[string]interface{}, error) {
	if x == nil {
		return nil, errors.New("map does not exist")
	}
	vals := make(map[string]interface{})
	for _, k := range x.Keys() {
		key, ok := skylark.AsString(k)
		if !ok {
			return nil, errors.New("key must be a string")
		}
		v, ok, err := x.Get(k)
		if err != nil {
			return nil, fmt.Errorf("couldn't find value a value for %q in the dict", key)
		}
		i, err := ValueToGo(v)
		if err != nil {
			return nil, errors.Wrap(err, "dict to map")
		}
		vals[key] = i
	}
	return vals, nil
}

func ListToTuple(x *skylark.List) (skylark.Tuple, error) {
	if x == nil {
		return nil, errors.New("list does not exist")
	}
	var vals skylark.Tuple
	var p skylark.Value
	iter := x.Iterate()
	defer iter.Done()
	for iter.Next(&p) {
		vals = append(vals, p)
	}
	return vals, nil
}

func ListToSlice(x *skylark.List) (interface{}, error) {
	if x == nil {
		return nil, errors.New("list does not exist")
	}
	var vals interface{}
	var p skylark.Value
	iter := x.Iterate()
	defer iter.Done()
	for iter.Next(&p) {
		v, err := ValueToGo(p)
		if err != nil {
			return nil, err
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
	case *skylark.Dict:
		return DictToMap(x)
	case *skylark.List:
		return ListToSlice(x)
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
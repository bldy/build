package skylark

import (
	"log"
	"reflect"

	"bldy.build/build"
	"bldy.build/build/internal"
	"bldy.build/build/label"
	"github.com/google/skylark"
	"github.com/pkg/errors"
)

func (s *skylarkVM) makeNativeRule(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	t := internal.Get(fn.Name()) // the reason why this rule is native is that it is registered as such for each of the native types like CC
	newReflectType := reflect.New(t)
	newStruct := newReflectType.Elem()
	for _, kwarg := range kwargs {
		strct, err := internal.GetFieldByTag(fn.Name(), string(kwarg.Index(0).(skylark.String)), t)

		if err != nil {
			log.Println(errors.Wrap(err, "make native rule"))
		}
		f := newStruct.FieldByName(strct.Name)
		v := kwarg.Index(1)
		switch f.Interface().(type) {
		case string:
			if s, ok := v.(skylark.String); ok {
				f.SetString(string(s))
			}
		case []string:
			if labels, ok := v.(*skylark.List); ok {
				var stringList []string
				i := labels.Iterate()
				var p skylark.Value
				for i.Next(&p) {
					if s, ok := p.(skylark.String); ok {
						stringList = append(stringList, string(s))
					}
				}
				f.Set(reflect.ValueOf(stringList))
			}
		case bool:
			if b, ok := v.(skylark.Bool); ok {
				f.SetBool(bool(b))
			}
		}
	}
	pkg := getPkg(thread)

	newNativeRule := newReflectType.Interface().(build.Rule)
	lbl := label.New(pkg, newNativeRule.Name())
	s.rules[lbl.String()] = newNativeRule
	return skylark.None, nil
}

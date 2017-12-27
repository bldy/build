package skylark

import (
	"log"
	"reflect"

	"bldy.build/build"
	"bldy.build/build/internal"
	"github.com/google/skylark"
)

func (s *skylarkVM) makeNativeRule(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	t := internal.Get(fn.Name()) // the reason why this rule is native is that it is registered as such
	newReflectType := reflect.New(t)
	newStruct := newReflectType.Elem()
	for _, kwarg := range kwargs {
		strct, err := internal.GasdetFieldByTag(fn.Name(), string(kwarg.Index(0).(skylark.String)), t)
		if err != nil {
			log.Println(err)
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
	s.rules = append(s.rules, newReflectType.Interface().(build.Rule))
	return skylark.None, nil
}

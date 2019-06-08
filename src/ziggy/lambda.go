package ziggy

import (
	"go.starlark.net/starlark"
)

/*
type lambda struct {
	name    string
	impl    *starlark.Function
	runtime build.Runtime
}

func (l *lambda) String() string        { panic("not implemented") }
func (l *lambda) Type() string          { return fmt.Sprintf("<stardust.lambda %q>", l.name) }
func (l *lambda) Freeze()               {}
func (l *lambda) Truth() starlark.Bool  { return starlark.True }
func (l *lambda) Hash() (uint32, error) { panic("not implemented") }
func (l *lambda) Name() string          { return l.name }

func (l *lambda) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name starlark.String

	if err := starlark.UnpackArgs(l.Name(), args, kwargs, ziggyKeyName, &name); err != nil {
		return nil, err
	}
	execContext := newContext(l.ctx, string(name))
	if val, err := starlark.Call(thread, l.impl, []starlark.Value{execContext}, nil); err != nil {
		return val, err
	}

	return execContext, nil
}
*/
func findArg(kw starlark.Value, kwargs []starlark.Tuple) starlark.Value {
	for i := 0; i < len(kwargs); i++ {
		if ok, err := starlark.Equal(kwargs[i].Index(0), kw); err == nil && ok {
			return kwargs[i].Index(1)
		} else if err != nil {
			return nil
		}
	}
	return nil
}

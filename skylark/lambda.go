package skylark

import (
	"fmt"
	"log"

	"github.com/google/skylark"
)

func (s *skylarkVM) makeRule(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var impl *skylark.Function
	attrs := new(skylark.Dict)
	err := skylark.UnpackArgs(fn.Name(), args, kwargs, skylarkKeyImpl, &impl, skylarkKeyAttrs, &attrs)
	if false && attrs != nil && err != nil {
		log.Println(err)
	}

	x := &lambdaFunc{
		skyFunc: impl,
		attrs:   attrs,
		vm:      s,
	}

	return x, nil
}

type lambdaFunc struct {
	name    string
	skyFunc *skylark.Function
	attrs   *skylark.Dict
	vm      *skylarkVM
}

func (l *lambdaFunc) Call(thread *skylark.Thread, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	return l.makeSkylarkRule(thread, args, kwargs)
}

func (l *lambdaFunc) Name() string          { return l.name }
func (l *lambdaFunc) Freeze()               {}
func (l *lambdaFunc) Truth() skylark.Bool   { return true }
func (l *lambdaFunc) String() string        { return fmt.Sprintf("name = %q", l.name) }
func (l *lambdaFunc) Hash() (uint32, error) { return 1337, nil }
func (l *lambdaFunc) Type() string          { return "lambda_func" }

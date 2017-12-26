package skylark

import (
	"fmt"
	"log"

	"github.com/google/skylark"
)

type lambdaFunc struct {
	name    string
	skyFunc *skylark.Function
	attrs   *skylark.Dict
	vm      *skylarkVM
}

func (l *lambdaFunc) Call(thread *skylark.Thread, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var name string
	var deps *skylark.List
	err := skylark.UnpackArgs(fmt.Sprintf("new rule (%s)", name), args, kwargs, skylarkKeyName, &name, skylarkKeyDeps, &deps)
	if deps != nil && err != nil {
		log.Println(err)
	}

	newRule := skylarkRule{
		name:      name,
		args:      args,
		kwargs:    kwargs,
		skyFunc:   l.skyFunc,
		skyThread: thread,
	}

	if deps != nil {
		i := deps.Iterate()
		var p skylark.Value
		for i.Next(&p) {
			if dep, ok := p.(skylark.String); ok {
				newRule.deps = append(newRule.deps, string(dep))
			}
		}
	}
	l.vm.rules = append(l.vm.rules, &newRule)
	return skylark.None, nil
}

func (l *lambdaFunc) Name() string          { return l.name }
func (l *lambdaFunc) Freeze()               {}
func (l *lambdaFunc) Truth() skylark.Bool   { return true }
func (l *lambdaFunc) String() string        { return fmt.Sprintf("name = %q", l.name) }
func (l *lambdaFunc) Hash() (uint32, error) { return 1337, nil }
func (l *lambdaFunc) Type() string          { return "lambda_func" }

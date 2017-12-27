package skylark

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"bldy.build/build/executor"
	"bldy.build/build/racy"
	"github.com/google/skylark"
	"github.com/pkg/errors"
)

type skylarkRule struct {
	Name         string
	Dependencies []string

	skyFuncLabel string
	skyThread    *skylark.Thread
	args         skylark.Tuple
	kwargs       []skylark.Tuple
	skyFunc      *skylark.Function
	attrs        *skylark.Dict
}

func (l *lambdaFunc) makeRule(thread *skylark.Thread, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var name string
	var deps *skylark.List
	err := skylark.UnpackArgs(fmt.Sprintf("new rule (%s)", name), args, kwargs, skylarkKeyName, &name, skylarkKeyDeps, &deps)
	if deps != nil && err != nil {
		log.Println(err)
	}

	newRule := skylarkRule{
		Name:         name,
		args:         args,
		kwargs:       kwargs,
		skyFunc:      l.skyFunc,
		skyThread:    thread,
		attrs:        l.attrs,
		skyFuncLabel: l.skyFunc.Name(),
	}
	newRule.WalkAttrs(func(kw skylark.Value, attr *attr) { // check the attributes
		if _, ok := findArg(kw, kwargs); !ok { // try finding the kwarg mentioned in the attribute
			if attr.def != nil { // if the attribute has a default and it's not in kwargs
				newRule.kwargs = append(newRule.kwargs, skylark.Tuple{kw, attr.def}) // set it
			}
		}
	})
	if deps != nil {
		i := deps.Iterate()
		var p skylark.Value
		for i.Next(&p) {
			if dep, ok := p.(skylark.String); ok {
				newRule.Dependencies = append(newRule.Dependencies, string(dep))
			}
		}
	}
	l.vm.rules = append(l.vm.rules, &newRule)
	return skylark.None, nil
}

func (s *skylarkRule) Build(e *executor.Executor) error {
	ctx := skylarkContext{
		label: s.skyFuncLabel,
		attrs: s.attrs,
	}
	thread := &skylark.Thread{}
	_, err := s.skyFunc.Call(thread, []skylark.Value{&ctx}, nil)
	return err
}

func (s *skylarkRule) WalkAttrs(wf walkAttrFunc) {
	if s.attrs != nil {
		i := s.attrs.Iterate()
		var p skylark.Value
		for i.Next(&p) {
			val, _, err := s.attrs.Get(p)
			if err != nil {
				log.Println(errors.Wrap(err, "skylarkRule walk attrs"))
			}
			if attr, ok := val.(*attr); ok {
				wf(p, attr)
			}
		}
	}
}

type walkAttrFunc func(skylark.Value, *attr)

func (s *skylarkRule) Hash() []byte {
	h := racy.New()
	io.WriteString(h, s.skyFuncLabel)
	funcHash, err := s.skyFunc.Hash()
	if err != nil {
		log.Fatal(err)
	}
	if err := binary.Write(h, binary.BigEndian, funcHash); err != nil {
		log.Fatal(err)
	}
	x := h.Sum(nil)
	s.WalkAttrs(walkAttrFunc(func(kw skylark.Value, attr *attr) {
		x = racy.XOR(x, s.hashArg(kw, attr))
	}))

	return x
}

func findArg(kw skylark.Value, kwargs []skylark.Tuple) (skylark.Value, bool) {
	for i := 0; i < len(kwargs); i++ {
		if ok, err := skylark.Equal(kwargs[i].Index(0), kw); err == nil && ok {
			return kwargs[i].Index(1), true
		} else if err != nil {
			return nil, false
		}
	}
	return nil, false
}

func (s *skylarkRule) hashArg(kw skylark.Value, attr *attr) []byte {
	h := racy.New()
	v, ok := findArg(kw, s.kwargs)
	if !ok {
		return nil
	}
	switch attr.t {
	case _int:
		if num, ok := v.(skylark.Int); ok {
			if i, ok := num.Int64(); ok {
				if err := binary.Write(h, binary.BigEndian, i); err != nil {
					log.Fatal(err)
				}
			}
		}
	case _labelList:
		if labels, ok := v.(*skylark.List); ok {
			var p skylark.Value
			i := labels.Iterate()
			for i.Next(&p) {
				if lbl, ok := p.(skylark.String); ok {
					io.WriteString(h, string(lbl))
				}
			}
		}
	}

	return h.Sum(nil)
}
func (s *skylarkRule) GetName() string {
	return s.Name
}

func (s *skylarkRule) GetDependencies() []string {
	return s.Dependencies
}
func (s *skylarkRule) Installs() map[string]string {
	return nil
}

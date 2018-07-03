package skylark

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
	"sevki.org/x/debug"

	"bldy.build/build/executor"
	"bldy.build/build/label"
	"bldy.build/build/racy"
	"github.com/google/skylark"
)

// Rule is a bazel rule that is implemented in skylark
type Rule struct {
	name string
	deps []label.Label

	SkyFuncLabel string
	skyThread    *skylark.Thread
	Args         skylark.Tuple
	Kwargs       []skylark.Tuple
	SkyFunc      *skylark.Function
	FuncAttrs    *skylark.Dict
	Attrs        nativeMap

	outputs []string

	Actions []executor.Action
}

func normalDeps(deps *skylark.List) ([]label.Label, error) {
	deplbls := []label.Label{}
	if deps != nil {
		i := deps.Iterate()
		var p skylark.Value
		for i.Next(&p) {
			if dep, ok := p.(skylark.String); ok {
				lbl, err := label.Parse(dep.String())
				if err != nil {
					return []label.Label{}, err
				}
				deplbls = append(deplbls, *lbl)
			}
		}
	}
	return deplbls, nil
}

func (f *lambdaFunc) makeSkylarkRule(thread *skylark.Thread, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	pkg := thread.Local(threadKeyPackage).(string)

	var name string
	var deps *skylark.List
	var outputs *skylark.List
	_ = outputs
	if n, ok := findArg(skylark.String(skylarkKeyName), kwargs); ok {
		if name, ok = skylark.AsString(n); !ok {
			return nil, errors.New("rule names have to be strings")
		}
	} else {
		return nil, errors.New("bldy rules need to have names")
	}
	if dps, ok := findArg(skylark.String(skylarkKeyDeps), kwargs); ok {
		if d, ok := dps.(*skylark.List); !ok {
			deps = d
		}
	}
	if outs, ok := findArg(skylark.String(skylarkKeyOutputs), kwargs); ok {
		if o, ok := outs.(*skylark.List); !ok {
			outputs = o
		}
	}
	lbl := label.Label{
		Name:    name,
		Package: &pkg,
	}

	ctx, outs, err := newContext(name, f.attrs, f.outputs, kwargs, f.vm.GetPackageDir(&lbl))
	if err != nil {
		return nil, errors.Wrap(err, "makeskylarkrule")
	}

	t := &skylark.Thread{
		Print: ctx.Print,
	}
	if _, err := f.skyFunc.Call(t, []skylark.Value{ctx}, nil); err != nil {
		return skylark.None, err
	}

	newRule := Rule{
		name:         name,
		Args:         args,
		Kwargs:       kwargs,
		SkyFunc:      f.skyFunc,
		skyThread:    thread,
		SkyFuncLabel: f.skyFunc.Name(),
		FuncAttrs:    f.attrs,
		Actions:      ctx.actionRecorder.calls,
		outputs:      outs,
	}

	if newRule.deps, err = normalDeps(deps); err != nil {
		debug.Println(err)
		return nil, errors.Wrap(err, "makeSkylarkRule")
	}

	f.vm.rules[lbl.String()] = &newRule

	return skylark.None, nil
}

// Build builds the skylarkRule
func (r *Rule) Build(e *executor.Executor) error {
	for _, action := range r.Actions {
		if err := action.Do(e); err != nil {
			panic(err)
			return err
		}
	}
	return nil

}

// Hash returns the calculated hash of a target
func (r *Rule) Hash() []byte {
	h := racy.New()
	io.WriteString(h, r.SkyFuncLabel)
	funcHash, err := r.SkyFunc.Hash()
	if err != nil {
		l.Fatal(err)
	}
	if err := binary.Write(h, binary.BigEndian, funcHash); err != nil {
		l.Fatal(err)
	}
	x := h.Sum(nil)
	WalkDict(r.FuncAttrs, func(kw skylark.Value, attr Attribute) error {
		x = racy.XOR(x, r.hashArg(kw, attr))
		return nil
	})
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

func (r *Rule) hashArg(kw skylark.Value, a Attribute) []byte {
	h := racy.New()
	v, ok := findArg(kw, r.Kwargs)
	if !ok {
		return nil
	}
	io.WriteString(h, v.String())
	return h.Sum(nil)
}

// GetName returns the name of the SkylarkRule
func (r *Rule) Name() string {
	return r.name
}

// GetDependencies returns the dependencies of the SkylarkRule
func (r *Rule) Dependencies() []label.Label {
	return r.deps
}

// Installs returns what will be outputed from the execution of the rule
func (r *Rule) Outputs() []string {
	return r.outputs
}

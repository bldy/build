package skylark

import (
	"encoding/binary"
	"fmt"
	"io"
	"path"

	"github.com/pkg/errors"

	"bldy.build/build/executor"
	"bldy.build/build/label"
	"bldy.build/build/racy"
	"github.com/google/skylark"
)

// Rule is a bazel rule that is implemented in skylark
type Rule struct {
	Name         string
	Dependencies []string

	SkyFuncLabel string
	skyThread    *skylark.Thread
	Args         skylark.Tuple
	Kwargs       []skylark.Tuple
	SkyFunc      *skylark.Function
	FuncAttrs    *skylark.Dict
	Attrs        nativeMap

	installs map[string]string

	Actions []executor.Action
}

func (f *lambdaFunc) makeSkylarkRule(thread *skylark.Thread, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var name string
	var deps *skylark.List
	err := skylark.UnpackArgs(fmt.Sprintf("new rule (%s)", name), args, kwargs, skylarkKeyName, &name, skylarkKeyDeps, &deps)
	// TODO(sevki): add debug mode here
	if false {
		l.Println(err)
	}
	lbl := label.Label{
		Name:    name,
		Package: label.Package(thread.Local(threadKeyPackage).(string)),
	}

	ctx, err := newContext(name, f.attrs, kwargs, f.vm.GetPackageDir(&lbl))
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
		Name:         name,
		Args:         args,
		Kwargs:       kwargs,
		SkyFunc:      f.skyFunc,
		skyThread:    thread,
		SkyFuncLabel: f.skyFunc.Name(),
		FuncAttrs:    f.attrs,
		Attrs:        make(nativeMap),
		Actions:      ctx.actionRecorder.calls,
		installs:     make(map[string]string),
	}
	outputs := skylark.StringDict{}
	ctx.outputs.ToStringDict(outputs)
	for _, file := range outputs {
		p := file.(*File).Path()
		_, file := path.Split(p)
		newRule.installs[path.Join("bin", file)] = p
	}

	if deps != nil {
		i := deps.Iterate()
		var p skylark.Value
		for i.Next(&p) {
			if dep, ok := p.(skylark.String); ok {
				newRule.Dependencies = append(newRule.Dependencies, string(dep))
			}
		}
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
func (r *Rule) GetName() string {
	return r.Name
}

// GetDependencies returns the dependencies of the SkylarkRule
func (r *Rule) GetDependencies() []string {
	return r.Dependencies
}

// Installs returns what will be outputed from the execution of the rule
func (r *Rule) Installs() map[string]string {
	return r.installs
}

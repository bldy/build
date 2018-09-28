package skylark

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"sevki.org/x/debug"

	"bldy.build/build/executor"
	"bldy.build/build/label"
	"bldy.build/build/racy"
	"bldy.build/build/workspace"
	"github.com/google/skylark"
)

// Rule is a bazel rule that is implemented in skylark
type Rule struct {
	name string
	deps []label.Label
	ws   workspace.Workspace

	SkyFuncLabel string
	skyThread    *skylark.Thread
	Args         skylark.Tuple
	KWArgs       []skylark.Tuple
	SkyFunc      *skylark.Function
	FuncAttrs    *skylark.Dict
	Attrs        *skylark.Dict

	host           label.Label
	compatibleWith []label.Label
	toolchains     []label.Label
	restrictedTo   []label.Label
	tags           []string
	outputs        []string
	files          []string
	Actions        []executor.Action

	ctx *context
}

func labelListToArray(labelList *skylark.List) ([]label.Label, error) {
	if labelList != nil {
		lbls := []label.Label{}
		i := labelList.Iterate()
		var p skylark.Value
		for i.Next(&p) {
			if l, ok := p.(label.Label); ok {
				if err := l.Valid(); err != nil {
					return nil, err
				}
				lbls = append(lbls, l)
			}
		}
		return lbls, nil
	}
	panic("list can't be nil")
}

func normalDeps(deps *skylark.List, rulepkg string) ([]label.Label, error) {
	deplbls := []label.Label{}
	if deps != nil {
		i := deps.Iterate()
		var p skylark.Value
		for i.Next(&p) {
			if dep, ok := p.(label.Label); ok {
				if err := dep.Valid(); err != nil {
					return nil, err
				}
				if !dep.IsAbs() {
					dep = label.New(rulepkg, dep.Name())
				}

				deplbls = append(deplbls, dep)
			}
		}
	}
	return deplbls, nil
}

func (f *lambdaFunc) makeSkylarkRule(thread *skylark.Thread, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	pkg := getPkg(thread)
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

	if outs, ok := findArg(skylark.String(skylarkKeyOutputs), kwargs); ok {
		if o, ok := outs.(*skylark.List); !ok {
			outputs = o
		}
	}
	lbl := label.New(pkg, name)

	ctx, skyio, err := newContext(name, f.attrs, f.outputs, kwargs, lbl, f.vm.ws)
	if err != nil {
		return nil, errors.Wrap(err, "makeskylarkrule")
	}

	t := &skylark.Thread{
		Print: ctx.Print,
	}

	if _, err := f.skyFunc.Call(t, []skylark.Value{ctx}, nil); err != nil {
		buf := bytes.NewBuffer(nil)
		thread.Caller().WriteBacktrace(buf)
		debug.Indent(buf, 1)
		return nil, fmt.Errorf("skylark: makeskylarkrule: call: %s\n%s", err.Error(), buf.String())
	}

	newRule := Rule{
		name:         name,
		Args:         args,
		KWArgs:       kwargs,
		SkyFunc:      f.skyFunc,
		skyThread:    thread,
		SkyFuncLabel: f.skyFunc.Name(),
		FuncAttrs:    f.attrs,
		Actions:      ctx.actionRecorder.calls,
		outputs:      skyio.outputs,
		files:        skyio.files,
		ctx:          ctx,
	}

	if dps, ok := ctx.attrs[skylarkKeyDeps]; ok {
		if d, ok := dps.(*skylark.List); ok {
			deps = d
		}
	}

	if newRule.compatibleWith, err = labelListToArray(ctx.attrs[skylarkKeyCompatibleWith].(*skylark.List)); err != nil {
		return nil, err
	}
	ok := false
	if newRule.host, ok = ctx.attrs[skylarkKeyHost].(label.Label); !ok {
		return nil, fmt.Errorf("host cannot be null, as it has a default value for all skylark rules")
	}
	if newRule.deps, err = normalDeps(deps, pkg); err != nil {
		return nil, errors.Wrap(err, "makeSkylarkRule.normalDeps")
	}
	f.vm.rules[lbl.String()] = &newRule
	return skylark.None, nil
}

// Build builds the skylarkRule
func (r *Rule) Build(e *executor.Executor) error {
	for _, action := range r.Actions {
		if err := action.Do(e); err != nil {
			return err
		}
	}
	return nil

}

func (r *Rule) Platform() label.Label          { return r.host }
func (r *Rule) Workspace() workspace.Workspace { return r.ws }

// Hash returns the calculated hash of a target
func (r *Rule) Hash() []byte {

	opts := []racy.Option{}
	for _, f := range r.files {
		opts = append(opts, racy.AllowExtension(filepath.Ext(f)))
	}
	h := racy.New(opts...)

	if err := h.HashFiles(r.files...); err != nil {
		panic(err)
	}

	io.WriteString(h, r.SkyFuncLabel)
	funcHash, err := r.SkyFunc.Hash()
	if err != nil {
		l.Fatal(err)
	}
	if err := binary.Write(h, binary.BigEndian, funcHash); err != nil {
		l.Fatal(err)
	}
	// sort Attributes
	keys := []string{}
	for k, _ := range r.ctx.attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v, ok := r.ctx.attrs[k].(skylark.Value)
		if ok {
			h.HashSkylarkValue(v)
		}
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
	v, ok := findArg(kw, r.KWArgs)
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

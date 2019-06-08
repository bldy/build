package ziggy

/*
// error(x) reports an error to the Go test framework.
func error_(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("error: got %d arguments, want 1", len(args))
	}
	buf := new(strings.Builder)
	thread.Caller().WriteBacktrace(buf)
	buf.WriteString("Error: ")
	if s, ok := starlark.AsString(args[0]); ok {
		buf.WriteString(s)
	} else {
		buf.WriteString(args[0].String())
	}

	return starlark.None, nil
}

var global = starlark.StringDict{
	"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
	"module": starlark.NewBuiltin("module", starlarkstruct.MakeModule),
	"error":  starlark.NewBuiltin("error", error_),
}

func httpLoader(u *url.URL, rt build.Runtime) (Package, error) {
	return &httpPackage{
		Dir:        u.Path,
		Name:       u.String(),
		BuildFiles: []string{u.String()},
		runtime:    rt,
		tasks:      make(map[string]*Task),
		wd:         wd,
	}, nil
}

type httpPackage struct {
	Dir  string // directory containing package sources
	Name string // package name

	// Source files
	BuildFiles []string // .bldy source files

	runtime build.Runtime

	tasks map[string]*Task

	wd string
}

func (pkg *httpPackage) Eval(thread *starlark.Thread) (starlark.StringDict, error) {
	if thread == nil {
		thread = &starlark.Thread{
			Name:  pkg.Name,
			Print: func(_ *starlark.Thread, msg string) { log.Println(msg) },
			Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
				u, err := url.Parse(module)
				if err != nil {
					return nil, err
				}
				log.Println(pretty.JSON(u))
				p, err := Load(u, pkg.ctx, pkg.wd)
				if err != nil {
					return nil, err
				}
				return p.Eval(thread)
			},
		}
	}
	predeclared := global
	predeclared["rule"] = starlark.NewBuiltin("rule", pkg.newRule)
	pkgVars := make(starlark.StringDict)
	for _, file := range pkg.BuildFiles {
		req, err := http.Get(file)
		if err != nil {
			return nil, errors.Wrap(err, "pkg eval")
		}
		local, err := starlark.ExecFile(thread, file, req.Body, predeclared)
		if err != nil {
			if evalErr, ok := err.(*starlark.EvalError); ok {
				log.Fatal(evalErr.Backtrace())
			}
			log.Fatal(err)
		}
		for k, v := range local {
			pkgVars[k] = v
		}

	}
	return pkgVars, nil
}

func (pkg *httpPackage) newRule(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var impl *starlark.Function
	attrs := new(starlark.Dict)
	outputs := new(starlark.Dict)

	if err := starlark.UnpackArgs("ziggy.newRule", args, kwargs, ziggyKeyImpl, &impl, ziggyKeyAttrs, &attrs, ziggyKeyOutputs, &outputs); err != nil {
		return nil, err
	}
	l := &lambda{
		impl: impl,
		ctx:  pkg.ctx,
	}

	return l, nil
}

const localKey = "Reporter"

func (pkg *httpPackage) GetTask(u *url.URL) (build.Task, error) {
	if rule, ok := pkg.tasks[u.Fragment]; ok {
		return rule, nil
	}
	return nil, fmt.Errorf("couldn't find rule %q", u)
}
*/
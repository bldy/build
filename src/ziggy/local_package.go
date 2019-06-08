package ziggy

/*
type localPackage struct {
	Dir  string // directory containing package sources
	Name string // package name

	// Source files
	BuildFiles []string // .bldy source files

	wd string

	ctx   build.Context
	u     url.URL
	tasks map[string]*Task
}

func (pkg *localPackage) Eval(thread *starlark.Thread) (starlark.StringDict, error) {
	if thread == nil {
		thread = &starlark.Thread{
			Name:  pkg.Name,
			Print: func(_ *starlark.Thread, msg string) { log.Println(msg) },
			Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
				u, err := url.Parse(module)
				if err != nil {
					return nil, err
				}
				p, err := Load(u, pkg.ctx, pkg.wd)
				if err != nil {
					return nil, err
				}
				return p.Eval(thread)
			},
		}
	}

	predeclared := make(starlark.StringDict)
	for k, v := range global {
		predeclared[k] = v
	}
	predeclared["rule"] = starlark.NewBuiltin("rule", pkg.makeRule)
	predeclared["export"] = starlark.NewBuiltin("export", pkg.export)
	pkgVars := make(starlark.StringDict)

	for _, file := range pkg.BuildFiles {
		f, err := os.Open(file)
		if err != nil {
			return nil, errors.Wrap(err, "pkg eval")
		}
		local, err := starlark.ExecFile(thread, file, f, predeclared)
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

func (pkg *localPackage) makeRule(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

func (pkg *localPackage) export(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	for _, v := range args {
		if ctx, ok := v.(*Context); ok {
			name := ctx.name
			u := pkg.u
			u.Fragment = name
			pkg.tasks[name] = &Task{
				name: u.String(),
				u:    u,
			}
			log.Println(name)
		} else {
			return nil, fmt.Errorf("was expecting exec context got %s instead", v.Type())
		}
	}
	return starlark.None, nil
}

func fileLoader(u *url.URL, bctx build.Context, wd string) (Package, error) {
	if u.Host == project.RootKey {
		rootdir, err := project.Search(wd, func(s string) (os.FileInfo, error) {
			for _, ext := range []string{".git"} {
				if fi, err := os.Stat(path.Join(s, ext)); err != os.ErrNotExist {
					return fi, nil
				}
			}
			return nil, os.ErrNotExist
		})
		if err != nil {
			return nil, err
		}
		u.Host = ""
		u.Path = path.Join(rootdir, u.Path)
	}
	dir, _ := filepath.Split(u.Path)
	files, err := filepath.Glob(filepath.Join(u.Path, "*.bldy"))
	if err != nil {
		return nil, errors.Wrap(err, "file loader")
	}
	return &localPackage{
		Dir:        u.Path,
		Name:       dir,
		BuildFiles: files,
		ctx:        bctx,
		tasks:      make(map[string]*Task),
		u:          *u,
		wd:         wd,
	}, nil
}

func (pkg *localPackage) GetTask(u *url.URL) (build.Task, error) {
	if err := pkg.absoluteURL(u); err != nil {
		return nil, errors.Wrap(err, "get target")
	}
	if rule, ok := pkg.tasks[u.Fragment]; ok {
		return rule, nil
	}
	return nil, fmt.Errorf("couldn't find rule %q", u)
}

func (pkg *localPackage) absoluteURL(u *url.URL) error {
	if u.Host == project.RootKey {
		u.Host = ""
		u.Path = path.Join(pkg.wd, u.Path)
		return nil
	}
	return nil
}
*/
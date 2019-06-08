package ziggy

/*
func newContext(ctx build.Context, name string) starlark.Value {
	c := &Context{
		ctx,
		name,
		exec.New(),
	}
	return c
}

type Context struct {
	build.Context
	name string
	exec exec.ActionModule
}

func (ctx *Context) String() string        { panic("not implemented") }
func (ctx *Context) Type() string          { return "build_context" }
func (ctx *Context) Freeze()               { panic("not implemented") }
func (ctx *Context) Truth() starlark.Bool  { panic("not implemented") }
func (ctx *Context) Hash() (uint32, error) { panic("not implemented") }
func (ctx *Context) AttrNames() []string   { panic("not implemented") }
func (ctx *Context) Attr(name string) (starlark.Value, error) {
	switch name {
	case "name":
		return starlark.String(ctx.name), nil
	case "os":
		return starlark.String(ctx.Context.OS()), nil
	case "arch":
		return starlark.String(ctx.Context.Arch()), nil
	case "exec":
		return &ctx.exec, nil
	default:
		return nil, fmt.Errorf("%q is not a ctx attribute", name)
	}
}
*/
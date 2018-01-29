"""Example of a rule that accesses its attributes."""

def _noop_impl(ctx):
	ctx.actions.do_nothing("nothing")

noop = rule(
    attrs = {},
    implementation = _noop_impl,
)

"""Example of a rule that accesses its attributes."""

def _impl(ctx):
  # Print debug information about the target.
  print("Target {}".format(ctx.label))


printer = rule(
    implementation=_impl,
    attrs={
        # Do not declare "name": It is added automatically.
        "number": attr.int(default = 1),
        "deps": attr.label_list(allow_files=True),
    })

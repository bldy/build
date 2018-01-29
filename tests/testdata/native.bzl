"""Example of a rule that accesses its attributes."""


cmd_deps = [
  "clang",
]

def _clang_binary_impl(ctx):
	cc_binary(
		name=ctx.attrs.name,
		copts=ctx.attrs.copts,
		includes=ctx.attrs.includes,
		deps=ctx.attrs.deps,
		strip=True,
		linkopts=ctx.attrs.linkopts,
	)

clang_binary = rule(
    attrs = {
        "srcs": attr.label_list(allow_files = True),
        "includes": attr.label_list(
            allow_files = True,
         ),
        "copts": attr.label_list(),
        "deps": attr.label_list(
            default = cmd_deps,
        ),
        "linkopts": attr.label_list(),
    },
    implementation = _clang_binary_impl,
)

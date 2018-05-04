"""Execute a binary.
The example below executes the binary target "//actions_run:merge" with
some arguments. The binary will be automatically built by Bazel.
The rule must declare its dependencies. To do that, we pass the target to
the attribute "_merge_tool". Since it starts with an underscore, it is private
and users cannot redefine it.
"""

def _impl(ctx):
	# The list of arguments we pass to the script.
	args = [ctx.outputs.out.path] + [f.path for f in ctx.files.srcs]
	print(args)

clang = rule(
    attrs = {
        "srcs": attr.label_list(allow_files = True),
        "out": attr.output(mandatory = True),
        "_clang": attr.label(
            executable = True,
            cfg = "host",
            allow_files = True,
        ),
    },
    implementation = _impl,
)

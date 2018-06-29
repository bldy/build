"""Execute a binary.
The example below executes the binary target "//actions_run:merge" with
some arguments. The binary will be automatically built by Bazel.
The rule must declare its dependencies. To do that, we pass the target to
the attribute "_merge_tool". Since it starts with an underscore, it is private
and users cannot redefine it.
"""

def _impl(ctx):
	# The list of arguments we pass to the script.
	args = [f.path for f in ctx.files.srcs] + ["-o"] + [ctx.outputs.out.path]  
	print(args)
  	ctx.actions.run(
 		arguments=args,
      		progress_message="Running: %s" % args,
		executable="clang")

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

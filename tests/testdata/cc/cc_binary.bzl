"""CC Binary is an example for compiling a c binary
"""

def _impl(ctx):
    # The list of arguments we pass to the script.
    args = [f.path for f in ctx.files.srcs] + ["-o"] + [ctx.outputs.out.path]
    print(args)
    ctx.actions.run(
        arguments = args,
        progress_message = "Running: %s" % args,
        executable = "clang",
    )

cc_binary = rule(
    attrs = {
        "srcs": attr.label_list(allow_files = True),
        "out": attr.output(mandatory = True),
        "deps": attr.label_list(),
    },
    implementation = _impl,
)
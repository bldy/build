"""CC Binary is an example for compiling a c binary
"""

def _impl(ctx):
    # The list of arguments we pass to the script.
    args = [f.path for f in ctx.files.srcs] + ["-o"] + [ctx.outputs.binary.path] + ["-l{}".format(d.name[3:]) for d in ctx.attrs.deps]
    print(args)
    ctx.actions.run(
        arguments = args,
        progress_message = "Running: %s" % args,
        executable = "/usr/bin/clang",
    )

cc_binary = rule(
    attrs = {
        "srcs": attr.label_list(allow_files = True),
        "deps": attr.label_list(allow_empty = True),
    },
    outputs = {"binary": "bin/%{name}"},
    implementation = _impl,
)
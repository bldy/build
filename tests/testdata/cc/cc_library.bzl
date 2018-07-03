"""CC library is an example for compiling a c library
"""

def _impl(ctx):
    # The list of arguments we pass to the script.
    args = ["-c"] + [f.path for f in ctx.files.srcs] + ["-o"] + [ctx.outputs.library.path]
    print(args)
    ctx.actions.run(
        arguments = args,
        progress_message = "Running: %s" % args,
        executable = "clang",
    )

cc_library = rule(
    attrs = {
        "srcs": attr.label_list(allow_files = True),
        "deps": attr.label_list(allow_empty = True),
    },
    outputs = {"library": "lib/%{name}.a"},
    implementation = _impl,
)
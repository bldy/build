load("testdata/empty.bzl", "empty")

empty(
	name="test",
	src=[
		"x.c",
	],
	deps=[
		"sevki",
	]
)


load("testdata/printer.bzl", "printer")

# Run 'bazel build :some_target' to print its information.
printer(
    name = "some_target",
    deps = [
         ":other",
    ],
)

# This target is a dependency of 'some_target', so it is automatically analyzed
# when you build 'some_target'.
printer(name = "other")

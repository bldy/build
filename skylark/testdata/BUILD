load("testdata/empty.bzl", "empty")

empty(
    name = "test",
    src = ["x.c"],
)

load("testdata/noop.bzl", "noop")

noop(
    name = "noop",
)

load("testdata/printer.bzl", "printer")

printer(
    name = "print",
    number = 1,
    deps = [":noop"],
)

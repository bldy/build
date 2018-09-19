load("cc_binary.bzl", "cc_binary")
load("cc_library.bzl", "cc_library")

cc_binary(
    name = "hello",
    srcs = [
        "hello.c",
    ],
)

cc_library(
    name = "libhello",
    srcs = ["libhello/hello.c"],
)

cc_binary(
    name = "hellowithlib",
    srcs = ["libhello/caller.c"],
    deps = [":libhello"],
)
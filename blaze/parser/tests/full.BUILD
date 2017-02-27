C_FLAGS = [
    "-Wall",
    "-ansi",
    "-Wno-unused-variable",
    "-pedantic",
    "-Werror",
    "-c",
]


CC_FLAGS = [
    "-Wall",
    "-ansi",
    "-Wno-unused-variable",
    "-pedantic",
    "-Werror",
    "-c"
]

XSTRING_SRCS = CC_FLAGS + C_FLAGS

cc_library(
    name="libxstring",
    hdrs=glob(["*.h"]),
    includes=[
        "/usr/lib/",
        "/usr/include"
    ],
    copts=C_FLAGS,
    srcs=XSTRING_SRCS,
)

cc_binary(
    name='test',
    srcs=[
        'tests/test.c',
    ],
    copts=C_FLAGS,
    deps=[
        ':libxstring',
    ]
)

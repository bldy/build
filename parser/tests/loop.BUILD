[genrule(
    name = "count_lines_" + f[:-3],  # strip ".cc"
    srcs = [f],
    outs = f[:-3],
    cmd = "wc -l $< >$@",
 ) for f in glob(["*_test.cc"])]
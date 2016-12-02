package main

import (
	"flag"
	"fmt"
	"os"

	_ "bldy.build/build/targets/build"
	"bldy.build/build/targets/cc"
	_ "bldy.build/build/targets/harvey"
	_ "bldy.build/build/targets/yacc"

	"bldy.build/build/builder"
)

var write = flag.Bool("w", false, "Write back?")

func usage() {
	fmt.Println(`usage:
	build fix [target]

Will fix all the c/c++ errors it can`)
	os.Exit(1)
}

func query(t string) {

	c := builder.New()

	if c.ProjectPath == "" {
		fmt.Fprintf(os.Stderr, "You need to be in a git project.\n\n")
		usage()
	}
	targ := c.Add(t).Target

	cflags := []string{"-c"}
	srcs := []string{}
	switch targ.(type) {
	case *cc.CBin:
		cbin := targ.(*cc.CBin)
		cflags = append(cflags, cbin.CompilerOptions...)
		cflags = append(cflags, cbin.Includes.Includes()...)

		srcs = cbin.Sources

	case *cc.CLib:
		clib := targ.(*cc.CLib)
		cflags = append(cflags, clib.CompilerOptions...)
		cflags = append(cflags, clib.Includes.Includes()...)
		srcs = clib.Sources
	}
	for _, src := range srcs {
		fixit(src, cflags, *write)
	}

}
func main() {

	flag.Parse()
	args := flag.Args()
	if flag.NArg() != 1 {
		usage()
	}
	query(args[0])
}

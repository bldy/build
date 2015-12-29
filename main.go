// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main // import "sevki.org/build"

import (
	"fmt"
	"log"
	"os"
	"time"

	"runtime"

	"github.com/fatih/color"

	"flag"

	"sevki.org/build/builder"
	_ "sevki.org/build/targets/cc"
	_ "sevki.org/build/targets/harvey"
	_ "sevki.org/build/targets/yacc"
	"sevki.org/build/term"
)

var (
	build = "version"
	usage = `usage: build target

We require that you run this application inside a git project.
All the targets are relative to the git project. 
If you are in a subfoler we will traverse the parent folders until we hit a .git file.
`
)
var (
	verbose = flag.Bool("v", false, "more verbose output")
)

func main() {
	flag.Parse()

	target := flag.Args()[0]
	if len(flag.Args()) < 1 {
		flag.Usage()
		printUsage()
	}
	switch target {
	case "version":
		version()
		return
	case "serve":
		target = flag.Args()[1]
		server(target)
	default:
		execute(target)
	}
}
func progress() {
	fmt.Println(runtime.NumCPU())
}
func printUsage() {
	fmt.Fprintf(os.Stderr, usage)
	os.Exit(1)

}
func version() {
	fmt.Printf("Build %s", build)
	os.Exit(0)
}
func doneMessage(s string) {
	fmt.Printf("[%s] %s\n", color.GreenString(" OK "), s)
}
func failMessage(s string) {
	fmt.Printf("[ %s ] %s\n", color.RedString("FAIL"), s)

}
func execute(t string) {
	c := builder.New()

	if c.ProjectPath == "" {
		fmt.Fprintf(os.Stderr, "You need to be in a git project.\n\n")
		printUsage()
	}
	c.Add(t)

	count := c.Total
	cpus := runtime.NumCPU()

	done := make(chan bool)

	go term.Listen(c.Updates, cpus, *verbose)
	go term.Run(done)

	go c.Execute(time.Second, cpus)
	for i := 0; i < count; i++ {
		select {
		case done := <-c.Done:
			if *verbose {
				doneMessage(done.GetName())
			}
		case err := <-c.Error:
			<-done
			log.Fatal(err)
			os.Exit(1)
		case <-c.Timeout:
			log.Println("your build has timed out")
		}

	}
	term.Exit()
	<-done
	os.Exit(1)
}

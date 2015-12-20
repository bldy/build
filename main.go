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

	"sevki.org/build/context"
	_ "sevki.org/build/targets/cc"
	_ "sevki.org/build/targets/harvey"
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

func main() {
	if len(os.Args) < 2 {
		printUsage()
	}
	target := os.Args[1]
	switch target {
	case "server":
		server()
	case "version":
		version()
		return
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
	c := context.New()

	if c.ProjectPath == "" {
		fmt.Fprintf(os.Stderr, "You need to be in a git project.\n\n")
		printUsage()
	}
	c.Parse(t)

	count := c.Total
	cpus := runtime.NumCPU()

	done := make(chan bool)

	go term.Listen(c.Updates, cpus, false)
	go term.Run(done)

	go c.Execute(time.Second, cpus)
	for i := 0; i < count; i++ {
		select {
		case <-c.Done:

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

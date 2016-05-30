// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main // import "sevki.org/build/cmd/build"

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"runtime"

	"flag"

	"sevki.org/build/builder"
 	_ "sevki.org/build/targets/build"
	_ "sevki.org/build/targets/cc"
	_ "sevki.org/build/targets/harvey"
 	_ "sevki.org/build/targets/yacc"
	"sevki.org/build/term"
	"sevki.org/lib/prettyprint"
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

	if len(flag.Args()) < 1 {
		flag.Usage()
		printUsage()
	}
	target := flag.Args()[0]
	switch target {
	case "version":
		version()
		return
	case "serve":
		target = flag.Args()[1]
		server(target)
	case "nuke":	
		os.RemoveAll("/tmp/build")
		if len(flag.Args()) >=2 {
			target = flag.Args()[1]
			execute(target)
		}
	case "query":
		target = flag.Args()[1]
		query(target)
	case "hash":
		target = flag.Args()[1]
		hash(target)
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
	fmt.Printf("[%s] %s\n", " OK ", s)
}
func failMessage(s string) {
	fmt.Printf("[ %s ] %s\n", "FAIL", s)

}
func hash(t string) {
	c := builder.New()

	if c.ProjectPath == "" {
		fmt.Fprintf(os.Stderr, "You need to be in a git project.\n\n")
		printUsage()
	}
	fmt.Printf("%x\n", c.Add(t).HashNode())
}

func query(t string) {

	c := builder.New()

	if c.ProjectPath == "" {
		fmt.Fprintf(os.Stderr, "You need to be in a git project.\n\n")
		printUsage()
	}
	fmt.Println(prettyprint.AsJSON(c.Add(t).Target))
}
func execute(t string) {

	c := builder.New()

	if c.ProjectPath == "" {
		fmt.Fprintf(os.Stderr, "You need to be in a git project.\n\n")
		printUsage()
	}
	c.Root = c.Add(t)
	c.Root.IsRoot = true

	if c.Root == nil {
		log.Fatal("We couldn't find the root")
	}

	cpus := int(float32(runtime.NumCPU()) * 1.25)

	done := make(chan bool)

	// If the app hangs, there is a log.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		f, _ := os.Create("/tmp/build-crash-log.json")
		fmt.Fprintf(f, prettyprint.AsJSON(c.Root))
		os.Exit(1)
	}()

	go term.Listen(c.Updates, cpus, *verbose)
	go term.Run(done)

	go c.Execute(time.Second, cpus)
	for {
		select {
		case done := <-c.Done:
			if *verbose {
				doneMessage(done.Url.String())
			}
			if done.IsRoot {
				goto FIN
			}
		case err := <-c.Error:
			<-done
			log.Fatal(err)
			os.Exit(1)
		case <-c.Timeout:
			log.Println("your build has timed out")
		}

	}
FIN:
	term.Exit()
	<-done
	os.Exit(0)
}

func compare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, c := range a {
		if c != b[i] {
			return false
		}
	}
	return true
}

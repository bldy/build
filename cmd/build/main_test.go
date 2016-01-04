// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main // import "sevki.org/build/cmd/build"

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"sevki.org/build/ast"
	"sevki.org/build/context"
)

//BUG(sevki): FOR TESTING PURPOSES THIS IS REALLY STUPID
func TestParse(t *testing.T) {
	wd, _ := os.Getwd()
	c := context.Context{
		Wd:          filepath.Join(wd, "tests/libxstring"),
		ProjectPath: filepath.Join(wd, "tests/libxstring"),
		Targets:     make(map[string]ast.Target),
	}

	if c.ProjectPath == "" {
		fmt.Fprintf(os.Stderr, "You need to be in a git project.\n\n")
		printUsage()
	}

	c.Parse("test")
}

func TestBuild(t *testing.T) {

	wd, _ := os.Getwd()
	c := context.Context{
		Wd:          filepath.Join(wd, "tests/libxstring"),
		ProjectPath: filepath.Join(wd, "tests/libxstring"),
		Error:       make(chan error),
		Timeout:     make(chan bool),
		Done:        make(chan ast.Target),
		BuildQueue:  make(chan *context.Node),
		Targets:     make(map[string]ast.Target),
	}

	if c.ProjectPath == "" {
		fmt.Fprintf(os.Stderr, "You need to be in a git project.\n\n")
		printUsage()
	}
	c.Parse("libxstring")

	go c.Execute(0, 3)

	for i := 0; i < c.Total; i++ {

		select {
		case done := <-c.Done:
			fmt.Printf("%d/%d :: %s\n", i+1, c.Total, done.GetName())
		case err := <-c.Error:
			failMessage(err.Error())
			t.Fail()
			continue
		case <-c.Timeout:
			log.Println("TIMEOUT")
			t.Fail()
			return
		}
	}
}

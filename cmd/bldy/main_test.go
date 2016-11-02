// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/bldy/build/builder"
)

func TestBuild(t *testing.T) {

	wd, _ := os.Getwd()
	c := builder.Builder{
		Wd:          filepath.Join(wd, "tests/libxstring"),
		ProjectPath: filepath.Join(wd, "tests/libxstring"),
		Error:       make(chan error),
		Timeout:     make(chan bool),
	}

	if c.ProjectPath == "" {
		fmt.Fprintf(os.Stderr, "You need to be in a git project.\n\n")
		printUsage()
	}

	go c.Execute(0, 3)

	for i := 0; i < c.Total; i++ {

		select {

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

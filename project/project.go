// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package project // import "bldy.build/build/project"

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"bldy.build/build/workspace"
	"github.com/vaughan0/go-ini"
)

var (
	file ini.File
	l    = log.New(os.Stdout, "project", 0)
)

func loadFile() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	s, err := workspace.FindWorkspace(wd, os.Stat)
	if err != nil {
		l.Println(err)
	}
	if file, err = ini.LoadFile(filepath.Join(s, "bldy.cfg")); err == nil {
		if err != nil {
			l.Printf("error: %v", err)
		}
	}
}

// Getenv returns the envinroment variable. It looks for the envinroment
// variable in the following order. It checks if the current shell session has
// an envinroment variable, checks if it's set in the OS specific section in
// the .build file, and checks it for common in the .build config file.
func Getenv(s string) string {
	if file == nil {
		loadFile()
	}
	if os.Getenv(s) != "" {
		return os.Getenv(s)
	} else if val, exists := file.Get(runtime.GOOS, s); exists {
		return val
	} else if val, exists := file.Get("", s); exists {
		return val
	}
	return ""
}

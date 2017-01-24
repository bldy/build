// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package project // import "bldy.build/build/project"

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/vaughan0/go-ini"
)

var (
	file ini.File

	pp = ""
)

func init() {
	wd, _ := os.Getwd()
	pp = GetGitDir(wd)
	var err error
	if file, err = ini.LoadFile(filepath.Join(Root(), "bldy.cfg")); err == nil {
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	}
}
func Root() (ProjectPath string) {
	return pp
}
func RelPPath(p string) string {
	rel, _ := filepath.Rel(Root(), p)
	return rel
}

func BuildOut() string {
	if os.Getenv("BUILD_OUT") != "" {
		return Getenv("BUILD_OUT")
	} else {
		return filepath.Join(
			Root(),
			"build_out",
		)
	}
}

func GetGitDir(p string) string {
	dirs := strings.Split(p, "/")
	for i := len(dirs) - 1; i > 0; i-- {
		try := fmt.Sprintf("/%s/.git", filepath.Join(dirs[0:i+1]...))
		if _, err := os.Lstat(try); os.IsNotExist(err) {
			continue
		} else if err != nil {
			log.Fatal(err)
		}
		pr, _ := filepath.Split(try)
		return pr
	}
	return ""
}

// Getenv returns the envinroment variable. It looks for the envinroment
// variable in the following order. It checks if the current shell session has
// an envinroment variable, checks if it's set in the OS specific section in
// the .build file, and checks it for common in the .build config file.
func Getenv(s string) string {
	if os.Getenv(s) != "" {
		return os.Getenv(s)
	} else if val, exists := file.Get(runtime.GOOS, s); exists {
		return val
	} else if val, exists := file.Get("", s); exists {
		return val
	} else {
		return ""
	}
}

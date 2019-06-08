// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package project // import "bldy.build/bldy/src/project"

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

var (
	l = log.New(os.Stdout, "project", 0)
)

const RootKey = "$PROJECTROOT"

// Getenv returns the envinroment variable. It looks for the envinroment
// variable in the following order. It checks if the current shell session has
// an envinroment variable, checks if it's set in the OS specific section in
// the .build file, and checks it for common in the .build config file.
func Getenv(s string) string {
	return os.Getenv(s)
}

// Search looks for recursively for a dir or file in each
// parent dir.
func Search(a string, stat Stat) (string, error) {
	dirs := strings.Split(a, "/")
	for i := len(dirs) - 1; i > 0; i-- {
		frags := append([]string{"/"}, dirs[0:i+1]...)
		path := path.Join(frags...)
		try := fmt.Sprintf("%s/.git", path)
		if _, err := stat(try); os.IsNotExist(err) {
			continue
		}
		return path, nil
	}
	return "", fmt.Errorf("workspace: new: %s is not a workspace", a)
}

// Stat checks if a file exists or not in a workspace
type Stat func(string) (os.FileInfo, error)

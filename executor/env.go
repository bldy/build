// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the https://github.com/golang/go/blob/master/LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package executor

import (
	"errors"
	"os/exec"

	"os"

	"path/filepath"

	"strings"
)

// ErrNotFound is the error resulting if a path search failed to find an executable file.
var ErrNotFound = errors.New("executable file not found in $PATH")

func findExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}

	if m := d.Mode(); !m.IsDir() && m&0111 != 0 {
		return nil
	}
	return os.ErrPermission
}

// lookPath searches for an executable binary named file
// in the directories named by the PATH environment variable.
// If file contains a slash, it is tried directly and the PATH is not consulted.
// The result may be an absolute path or a path relative to the current directory.
func lookPath(file string, env []string) (string, error) {
	// NOTE(rsc): I wish we could use the Plan 9 behavior here
	// (only bypass the path if file begins with / or ./ or ../)
	// but that would not match all the Unix shells.
	if strings.Contains(file, "/") {
		err := findExecutable(file)
		if err == nil {
			return file, nil
		}
		return "", &exec.Error{file, err}
	}
	path := os.Getenv("PATH")
	for _, p := range env {
		if pth := strings.TrimLeft(p, "PATH="); pth != p {
			path = p
		}
	}

	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}

		path := filepath.Join(dir, file)
		if err := findExecutable(path); err == nil {
			return path, nil
		}
	}
	return "", &exec.Error{file, ErrNotFound}
}

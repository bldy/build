// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"io"
	"os/exec"
	"sync"
)

// Exec executes a command with and writes to stdout and stderr
// outputs without combining the two streams.
func Exec(o, e io.Writer, cmd string, env, params []string) error {
	var stdOut, stdErr io.ReadCloser
	var wg sync.WaitGroup

	x := exec.Command(cmd, params...)

	stdErr, err := x.StderrPipe()
	if err != nil {
		return err
	}
	stdOut, err = x.StdoutPipe()
	if err != nil {
		return err
	}

	wg.Add(2)

	go func() {
		io.Copy(o, stdOut)
		wg.Done()
	}()

	go func() {
		io.Copy(e, stdErr)
		wg.Done()
	}()

	if err := x.Run(); err != nil {
		return err
	}

	wg.Wait()
	return nil
}

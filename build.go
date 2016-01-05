// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package build defines build target and build context structures
package build // import "sevki.org/build"
import (
	"bytes"
	"io"
	"log"
	"os/exec"
	"sync"
)

// Target is for implementing different build targets.
type Target interface {
	Build(*Context) error
	Installs() map[string]string
	GetName() string
	GetDependencies() []string
	Hash() []byte
}

// Context defines the context in which a target will be built, it
// provide helper functions for shelling out without having to worry
// about stdout or stderr outputs.
type Context struct {
	wd             string
	stderr, stdout *bytes.Buffer
	logger         *log.Logger
	buf            *bytes.Buffer
}

// NewContext initializes and returns a new build.Context
func NewContext(dir string) *Context {
	buf := bytes.Buffer{}
	return &Context{
		wd:     dir,
		stderr: &buf,
		stdout: &buf,
		logger: log.New(&buf, "", log.Lmicroseconds),
		buf:    &buf,
	}
}
func (c *Context) Stdout() io.Reader {
	return c.buf
}

func (c *Context) StdErr() io.Reader {
	return c.buf
}

func (c *Context) Printf(format string, v ...interface{}) {
	c.logger.Printf(format, v)
}

func (c *Context) Println(v ...interface{}) {
	c.logger.Println(v)
}

// Exec executes a command writing it's outputs to the context
func (c *Context) Exec(cmd string, env, params []string) error {
	var stdOut, stdErr io.ReadCloser
	var wg sync.WaitGroup

	x := exec.Command(cmd, params...)
	x.Dir = c.wd
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
		io.Copy(c.stdout, stdOut)
		wg.Done()
	}()

	go func() {
		io.Copy(c.stderr, stdErr)
		wg.Done()
	}()

	if err := x.Run(); err != nil {
		return err
	}

	wg.Wait()
	return nil
}

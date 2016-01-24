// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package build defines build target and build context structures
package build // import "sevki.org/build"
import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
	"sevki.org/build/util"
)

var vars map[string]string

func init() {
	if data, err := ioutil.ReadFile(filepath.Join(util.GetProjectPath(), ".build")); err == nil {
		vars = make(map[string]string)

		err = yaml.Unmarshal(data, &vars)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	}
}

func Getenv(s string) string {
	if val, exists := vars[s]; exists {
		return val
	} else {
		return os.Getenv(s)
	}
}

// Target defines the interface that rules must implement for becoming build targets.
type Target interface {
	GetName() string
	GetDependencies() []string

	Hash() []byte
	Build(*Context) error
	Installs() map[string]string
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
	c.Println(strings.Join(append([]string{cmd}, params...), "\n"))
	var stdOut, stdErr io.ReadCloser
	var wg sync.WaitGroup

	x := exec.Command(cmd, params...)
	x.Dir = c.wd
	x.Env = env
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

// Create creates and returns a new file with the given name in the context
func (c *Context) Create(name string) (*os.File, error) {
	return os.Create(filepath.Join(c.wd, name))
}

// Create creates and returns a new file with the given name in the context
func (c *Context) Open(name string) (*os.File, error) {
	return os.Open(filepath.Join(c.wd, name))
}

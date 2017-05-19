// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package build defines build target and build context structures
package build 

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"bldy.build/build/url"
)

var (
	printenv = flag.Bool("e", false, "prints envinroment variables into the log")
)

// Target defines the interface that rules must implement for becoming build targets.
type Target interface {
	GetName() string
	GetDependencies() []string

	Hash() []byte
	Build(*Runner) error
	Installs() map[string]string
}

// Runner defines the envinroment in which a target will be built, it
// provide helper functions for shelling out without having to worry
// about stdout or stderr outputs.
type Runner struct {
	ctx context.Context
	wd  string
	run []*Run
	log []fmt.Stringer
}

func (r *Runner) Context() context.Context {
	return r.ctx
}

type Run struct {
	At     time.Time
	Cmd    string
	Args   []string
	Output []byte
	Env    []string
	Err    error
}

func (r *Run) String() string {
	buf := bytes.Buffer{}
	if *printenv {
		buf.WriteString("envinroment variables: " + strings.Join(r.Env, "\n"))
	}
	buf.WriteString(strings.Join(append([]string{r.Cmd}, r.Args...), "\n"))

	buf.WriteString("\n")
	buf.Write(r.Output)
	return string(buf.String())
}

type Message string

func (m Message) String() string {
	return string(m)
}
func (r *Runner) RunCmds() []*Run {
	return r.run
}

func (r *Runner) Log() []fmt.Stringer {
	return r.log
}

// NewContext initializes and returns a new build.Context
func NewRunner(ctx context.Context, dir string) *Runner {
	return &Runner{
		wd:  dir,
		ctx: ctx,
	}
}

func (r *Runner) Printf(format string, v ...interface{}) {
	r.log = append(r.log, Message(fmt.Sprintf(format, v...)))
}

func (r *Runner) Println(v ...interface{}) {
	r.log = append(r.log, Message(fmt.Sprintln(v...)))
}

// Exec executes a command writing it's outputs to the context
func (r *Runner) Exec(cmd string, env, args []string) error {
	run := Run{
		At:   time.Now(),
		Cmd:  cmd,
		Args: args,
		Env:  env,
	}
	x := exec.CommandContext(r.ctx, cmd, args...)
	x.Dir = r.wd
	x.Env = env
	run.Output, run.Err = x.CombinedOutput()
	r.run = append(r.run, &run)
	r.log = append(r.log, &run)
	return run.Err
}

// Run executes a command writing it's outputs to the context
func (r *Runner) Run(ctx context.Context, cmd string, env, params []string) *exec.Cmd {
	x := exec.CommandContext(ctx, cmd, params...)

	x.Dir = r.wd
	x.Env = env
	return x
}

// Create creates and returns a new file with the given name in the context
func (r *Runner) Create(name string) (*os.File, error) {
	return os.Create(filepath.Join(r.wd, name))
}

// Open creates and returns a new file with the given name in the context
func (r *Runner) Open(name string) (*os.File, error) {
	if filepath.IsAbs(name) {
		return os.Open(name)
	}
	return os.Open(filepath.Join(r.wd, name))
}

func (r *Runner) Mkdir(name string) error {
	return os.MkdirAll(filepath.Join(r.wd, name), os.ModeDir|os.ModePerm)
}

// VM seperate the parsing and evauluating targets logic from rest of bldy
// so we can implement and use new grammars like jsonnet or go it self.
type VM interface {
	GetTarget(url.URL) (Target, error)
}

// Status represents a nodes status.
type Status int

const (
	Success Status = iota
	Fail
	Pending
	Started
	Fatal
	Warning
	Building
)

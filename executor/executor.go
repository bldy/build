// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor // import "bldy.build/build/executor"
import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	printenv = flag.Bool("e", false, "prints envinroment variables into the log")
)

// Action interface is used for deferred actions that get performed
// during the build stage, unlike rules actions are NOT meant to be executed
// in parralel.
type Action interface {
	Do(*Executor) error
}

// Executor defines the envinroment in which a target will be built, it
// provide helper functions for shelling out without having to worry
// about stdout or stderr outputs.
type Executor struct {
	ctx context.Context
	wd  string
	run []*Run
	log []fmt.Stringer
}

// Context returns the context that's attached to the Executor
func (e *Executor) Context() context.Context {
	return e.ctx
}

// Run defines the execution step
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

// Message defines a log Item in the message
type Message string

func (m Message) String() string {
	return string(m)
}

// RunCmds commands returns the commands that ran
func (e *Executor) RunCmds() []*Run {
	return e.run
}

// Log returns the logs
func (e *Executor) Log() []fmt.Stringer {
	return e.log
}

// New initializes and returns a new executor.Executor
func New(ctx context.Context, dir string) *Executor {
	return &Executor{
		wd:  dir,
		ctx: ctx,
	}
}

// Printf wraps sprintf for log items
func (e *Executor) Printf(format string, v ...interface{}) {
	e.log = append(e.log, Message(fmt.Sprintf(format, v...)))
}

// Println wraps sprintf for log items
func (e *Executor) Println(v ...interface{}) {
	e.log = append(e.log, Message(fmt.Sprintln(v...)))
}

func (e *Executor) CombinedLog() string {
	log := e.Log()
	buf := bytes.Buffer{}
	for _, logEntry := range log {
		buf.WriteString(logEntry.String())
	}
	return buf.String()
}

// Exec executes a command writing it's outputs to the context
func (e *Executor) Exec(cmd string, env, args []string) error {
	run := Run{
		At:   time.Now(),
		Cmd:  cmd,
		Args: args,
		Env:  env,
	}
	x := exec.CommandContext(e.ctx, cmd, args...)
	x.Dir = e.wd
	x.Env = env
	run.Output, run.Err = x.CombinedOutput()
	e.run = append(e.run, &run)
	e.log = append(e.log, &run)
	if run.Err != nil {
		return errors.New(string(run.Output))
	}
	return run.Err
}

// Run executes a command writing it's outputs to the context
func (e *Executor) Run(ctx context.Context, cmd string, env, params []string) *exec.Cmd {
	x := exec.CommandContext(ctx, cmd, params...)
	x.Dir = e.wd
	x.Env = env
	return x
}

// Create creates and returns a new file with the given name in the context
func (e *Executor) Create(name string) (*os.File, error) {
	return os.Create(filepath.Join(e.wd, name))
}

// Open creates and returns a new file with the given name in the context
func (e *Executor) Open(name string) (*os.File, error) {
	if filepath.IsAbs(name) {
		return os.Open(name)
	}
	return os.Open(filepath.Join(e.wd, name))
}

// Mkdir creates a folder in the executor
func (e *Executor) Mkdir(name string) error {
	return os.MkdirAll(filepath.Join(e.wd, name), os.ModeDir|os.ModePerm)
}

// Status represents a nodes status.
type Status int

const (
	// Success is success
	Success Status = iota
	// Fail is a failed job
	Fail
	// Pending is a pending job
	Pending
	// Started is a started job
	Started
	// Fatal is a fatal crash
	Fatal
	// Warning is a job that has warnings
	Warning
	// Building is a job that's being built
	Building
)

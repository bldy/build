// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor // import "bldy.build/build/executor"
import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"bldy.build/build/namespace"
	"sevki.org/x/debug"
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
	ns  namespace.Namespace
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
func New(ctx context.Context, ns namespace.Namespace) *Executor {
	return &Executor{
		ns:  ns,
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

func (e *Executor) env() []string {
	return os.Environ()
}

// Exec executes a command writing it's outputs to the context
func (e *Executor) Exec(cmd string, env, args []string) error {
	env = append(e.env(), env...)

	run := Run{
		At:   time.Now(),
		Cmd:  cmd,
		Args: args,
		Env:  env,
	}

	x := e.ns.Cmd(e.ctx, cmd, args...)

	run.Output, run.Err = x.CombinedOutput()
	envbuf := bytes.NewBufferString(strings.Join(env, "\n"))
	e.run = append(e.run, &run)
	e.log = append(e.log, &run)
	if run.Err != nil {
		errbuf := bytes.NewBuffer(run.Output)
		debug.Indent(errbuf, 2)
		debug.Indent(envbuf, 2)
		return fmt.Errorf(`%s
	command: %s
 	emv: 
%s
	output:
%s
`,
			run.Err,
			append([]string{cmd}, args...),
			envbuf.String(),
			errbuf.String(),
		)
	}
	return run.Err
}

// Run executes a command writing it's outputs to the namespace
func (e *Executor) Run(ctx context.Context, cmd string, args ...string) namespace.Cmd {
	return e.ns.Cmd(e.ctx, cmd, args...)

}

// Create creates and returns a new file with the given name in the namespace
func (e *Executor) Create(name string) (*os.File, error) {
	return e.ns.Create(name)
}

// OpenFile creates and returns a new file with the given name, flags and mode in the namespace
func (e *Executor) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return e.ns.OpenFile(name, flag, perm)
}

// Open creates and returns a new file with the given name in the namespace
func (e *Executor) Open(name string) (*os.File, error) {
	return e.ns.Open(name)
}

// Mkdir creates a folder in the executor
func (e *Executor) Mkdir(name string) error {
	return e.ns.Mkdir(name)
}

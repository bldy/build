// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package build defines build target and build context structures
package build

import (
	"bldy.build/build/executor"
	"bldy.build/build/label"
	"bldy.build/build/workspace"
)

//go:generate stringer -type=Status
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

var (
	HostPlatform    = label.Label("@bldy//platforms:host")
	DefaultPlatform = HostPlatform
)

// Rule defines the interface that rules must implement for becoming build targets.
type Rule interface {
	Name() string
	Dependencies() []label.Label
	Outputs() []string
	Hash() []byte
	Build(*executor.Executor) error
	Platform() label.Label
	Workspace() workspace.Workspace
}

// VM seperate the parsing and evauluating targets logic from rest of bldy
// so we can implement and use new grammars like jsonnet or go it self.
type VM interface {
	GetTarget(label.Label) (Rule, error)
}

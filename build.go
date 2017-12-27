// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package build defines build target and build context structures
package build

import (
	"bldy.build/build/executor"
	"bldy.build/build/label"
)

// Rule defines the interface that rules must implement for becoming build targets.
type Rule interface {
	GetName() string
	GetDependencies() []string
	Hash() []byte
	Build(*executor.Executor) error
	Installs() map[string]string
}

// VM seperate the parsing and evauluating targets logic from rest of bldy
// so we can implement and use new grammars like jsonnet or go it self.
type VM interface {
	GetTarget(*label.Label) (Rule, error)
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

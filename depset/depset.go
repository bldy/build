// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package depset

import (
	"path/filepath"

	"bldy.build/build"
	"bldy.build/build/executor"
	"bldy.build/build/label"
	"bldy.build/build/workspace"
)

type Depset struct {
	name         string        `group:"name"`
	dependencies []label.Label `group:"deps"`
	outputs      []string
	Prefix       *string `group:"prefix" group:"prefix" build:"expand"`
}

func (d *Depset) Hash() []byte {
	return []byte(d.name)
}

func (d *Depset) Build(e *executor.Executor) error {
	return nil
}

func (d *Depset) Name() string {
	return d.name
}

func (d *Depset) Dependencies() []label.Label {
	return d.dependencies
}
func (d *Depset) Outputs() []string {
	if d.Prefix != nil {
		outs := []string{}
		for _, o := range d.outputs {
			outs = append(outs, filepath.Join(*d.Prefix, o))
		}
		return outs
	}
	return d.outputs
}
func (d *Depset) AddOutput(output string) {
	d.outputs = append(d.outputs, output)
}

func (d *Depset) Platform() label.Label          { return build.DefaultPlatform }
func (d *Depset) Workspace() workspace.Workspace { return nil }

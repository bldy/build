// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package internal is used for registering types in build, it had no clear place
// in other packages to go which is why it gets it's own package
package internal

import (
	"bldy.build/bldy/src/build"
)

var (
	vms = make(map[string]build.VM)
)

// Register function is used to register new types of targets.
func Register(name string, vm build.VM) error {

	vms[name] = vm

	return nil
}

// Get returns a reflect.Type for a given name.
func Get(name string) build.VM {
	if t, ok := vms[name]; ok {
		return t
	}
	return nil
}

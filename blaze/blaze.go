// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package blaze

import (
	"fmt"
	"log"
	"os"

	"bldy.build/build/url"

	"bldy.build/build"

	"bldy.build/build/blaze/processor"
)

var l = log.New(os.Stdout, "blaze: ", 0)

type VM struct {
	wd      string
	targets map[string]map[string]build.Target
}

func NewVM(wd string) *VM {
	return &VM{
		targets: make(map[string]map[string]build.Target),
		wd:      wd,
	}
}
func (vm *VM) GetTarget(u url.URL) (build.Target, error) {
	if t, ok := vm.targets[u.Package][u.Target]; ok {
		return t, nil
	}

	p, err := processor.NewProcessorFromURL(u, vm.wd)
	if err != nil {
		l.Fatal(err)
	}
	go p.Run()
	// bug(sevki): this is a really bad way of doing this, there should me
	// some caching mechanism for this, it is yet to come !!
	for t := <-p.Targets; t != nil; t = <-p.Targets {
		if vm.targets[u.Package] == nil {
			vm.targets[u.Package] = make(map[string]build.Target)
		}
		vm.targets[u.Package][t.GetName()] = t
	}
	if t, ok := vm.targets[u.Package][u.Target]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("blaze vm: couldn't find target %q in package %q", u.Target, u.Package)
}
// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harvey

import (
	"crypto/sha1"
	"fmt"
	"io"
	"path/filepath"

	"bldy.build/build"
	"bldy.build/build/project"
)

type ObjCopy struct {
	Name         string   `objcopy:"name"`
	Dependencies []string `objcopy:"deps"`
	In           string   `objcopy:"infile"`
	Out          string   `objcopy:"outfile"`
}

func (oc *ObjCopy) GetName() string {
	return oc.Name
}

func (oc *ObjCopy) GetDependencies() []string {
	return oc.Dependencies
}

func (oc *ObjCopy) Hash() []byte {
	h := sha1.New()
	io.WriteString(h, oc.In)
	io.WriteString(h, oc.Out)
	io.WriteString(h, oc.Name)
	return []byte{}
}

// Had to be done
func Copier() string {
	if tpfx := project.Getenv("TOOLPREFIX"); tpfx == "" {
		return "objcopy"
	} else {
		return fmt.Sprintf("%s%s", tpfx, "objcopy")
	}
}
func (oc *ObjCopy) Build(e *build.Executor) error {
	params := []string{}
	params = append(params, "-I")
	params = append(params, oc.In)
	params = append(params, "-O")
	params = append(params, oc.Out)
	params = append(params, filepath.Join("bin", split(oc.Dependencies[0], ":")))
	params = append(params, oc.Name)
	if err := e.Exec(Copier(), nil, params); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}
func (oc *ObjCopy) Installs() map[string]string {
	installs := make(map[string]string)
	installs[oc.Name] = oc.Name
	return installs
}

// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harvey

import (
	"io"
	"os/exec"

	"bldy.build/build/executor"
	"bldy.build/build/racy"
)

type Sed struct {
	Name         string   `sed:"name"`
	Dependencies []string `sed:"deps"`
	Args         []string `sed:"args"`
	File         string   `sed:"file" build:"path"`
	Script       string   `sed:"script" build:"path"`
}

func (s *Sed) GetName() string {
	return s.Name
}

func (s *Sed) GetDependencies() []string {
	return s.Dependencies
}

func (s *Sed) Hash() []byte {
	h := racy.New()

	io.WriteString(h, s.Name)
	return []byte{}
}

func (s *Sed) Build(e *executor.Executor) error {
	params := s.Args
	if s.Script != "" {
		params = append(params, "-f", s.Script)
	}
	params = append(params, s.File)
	out, err := exec.Command("sed", params...).Output()
	if err != nil {
		return err
	}
	f, err := e.Create(s.Name)
	if err != nil {
		return err
	}

	f.Write(out)
	return nil
}
func (s *Sed) Installs() map[string]string {
	installs := make(map[string]string)
	installs[s.Name] = s.Name
	return installs
}

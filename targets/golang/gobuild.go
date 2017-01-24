// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package golang

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"bldy.build/build"
	"bldy.build/build/racy"
)

type GoBuild struct {
	Name         string   `go_build:"name"`
	Dependencies []string `go_build:"deps"`
	Sources      []string `go_build:"srcs" go_build:"srcs" build:"path"`
}

func (g *GoBuild) GetName() string {
	return g.Name
}

func (g *GoBuild) GetDependencies() []string {
	return g.Dependencies
}

func (g *GoBuild) Hash() []byte {
	h := sha1.New()
	io.WriteString(h, gover)
	io.WriteString(h, g.Name)
	racy.HashFiles(h, []string(g.Sources))
	return h.Sum(nil)
}

func (g *GoBuild) Build(c *build.Context) error {
	if err := c.Mkdir("_obj/exe"); err != nil {
		return fmt.Errorf("go mkdir: %s", err.Error())
	}
	cParams := []string{
		"-o",
		fmt.Sprintf("%s.a", g.Name),
		"-p", "main", "-complete", "-buildid", fmt.Sprintf("%x", g.Hash()),
		"-I", filepath.Join(os.Getenv("GOPATH"), "pkg", "linux_amd64"),
		"-pack",
	}
	cParams = append(cParams, g.Sources...)
	if err := c.Exec("/home/sevki/code/golang/pkg/tool/linux_amd64/compile", nil, cParams); err != nil {
		return fmt.Errorf("go compile: %s", err.Error())
	}

	lParams := []string{
		"-o",
		"_obj/exe/a.out",
		"-L", filepath.Join(os.Getenv("GOPATH"), "pkg", "linux_amd64"),
		"-extld=gcc",
		"-buildmode=exe",
		"-buildid", fmt.Sprintf("%x", g.Hash()),
		fmt.Sprintf("%s.a", g.Name),
	}
	if err := c.Exec("/home/sevki/code/golang/pkg/tool/linux_amd64/link", nil, lParams); err != nil {
		return fmt.Errorf("go link: %s", err.Error())
	}
	return nil
}

func (g *GoBuild) Installs() map[string]string {
	exports := make(map[string]string)

	exports[filepath.Join("bin", g.Name)] = "_obj/exe/a.out"

	return exports
}

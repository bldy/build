// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder // import "sevki.org/build/builder"

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"strings"

	"runtime"

	"sevki.org/build/ast"
	_ "sevki.org/build/targets/cc"
	"sevki.org/lib/prettyprint"
)

func TestBuild(t *testing.T) {

	wd, _ := os.Getwd()
	wds := strings.Split(wd, "/")
	wds = wds[:len(wds)-1]
	wds = append(wds, "tests", "libxstring")
	c := Context{
		Wd:          "/" + filepath.Join(wds...),
		ProjectPath: "/" + filepath.Join(wds...),
		Targets:     make(map[string]build.Target),
	}

	c.Parse("test")
	c.Execute(0, runtime.NumCPU())
	log.Printf(prettyprint.AsJSON(c.Root))

}

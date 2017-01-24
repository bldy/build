// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package processor // import "bldy.build/build/processor"
import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"

	"bldy.build/build/ast"
	"bldy.build/build/project"
)

func (p *Processor) glob(f *ast.Func) []string {
	wd := p.parser.Path
	if !filepath.IsAbs(wd) {
		log.Fatalf("Error parsing glob: %s is not an absolute path.", wd)
	}

	var files []string
	var excludes []*regexp.Regexp

	if len(f.AnonParams) != 1 {
		return []string{"Error parsing glob: proper usage is like so glob(include, exclude=[], exclude_directories=1)"}
	}

	if exs, ok := f.Params["exclude"]; ok {

		for _, ex := range exs.([]interface{}) {
			r, _ := regexp.Compile(ex.(string))
			excludes = append(excludes, r)
		}
	}

	//BUG(sevki): put some type checking here
	for _, s := range f.AnonParams[0].([]interface{}) {
		globPtrn := ""

		switch s.(type) {
		case string:
			globPtrn = filepath.Clean(filepath.Join(wd, s.(string)))
			log.Println(globPtrn)
		default:
			return nil
		}

		globFiles, err := filepath.Glob(globPtrn)

		if err != nil {
			return []string{"Error parsing glob: %s"}
		}

		for _, f := range globFiles {
			f, _ := filepath.Rel(project.Root(), f)
			f = fmt.Sprintf("//%s", f)
		}
	RESIZED:
		for i, f := range globFiles {
			for _, x := range excludes {
				if x.Match([]byte(f)) {
					globFiles = append(globFiles[:i], globFiles[i+1:]...)
					goto RESIZED
				}
			}
			globFiles[i] = f
		}

		files = append(files, globFiles...)
	}
	return files
}

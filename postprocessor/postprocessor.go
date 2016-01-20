// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package builder parses build graphs and coordinates builds
package postprocessor // import "sevki.org/build/postprocessor"
import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"strings"

	"sevki.org/build"
	"sevki.org/build/util"
)

type PostProcessor struct {
	projectPath, packagePath string
}

// New returns a new PostProcessor
func New(p string) PostProcessor {
	return PostProcessor{
		packagePath: p,
		projectPath: util.GetProjectPath(),
	}
}

// ProcessDependencies takes relative dependency paths and turns then in to
// absolute paths.
func (pp *PostProcessor) ProcessDependencies(t build.Target) error {

	v := reflect.ValueOf(t)

	deps := v.Elem().FieldByName("Dependencies").Interface().([]string)

	seen := make(map[string]bool)

	for i, d := range deps {
		if _, ok := seen[d]; ok {
			log.Fatalf("post process dependencies: %s is duplicated", d)
		}
		switch {
		case d[:2] == "//":
			continue
		case d[0] == ':':
			deps[i] = fmt.Sprintf("//%s%s", pp.packagePath, d)
			break
		default:
			errorf := `dependency '%s' in %s is not a valid URL for a target.
		a target url can only start with a '//' or a ':' for relative targets.`

			return fmt.Errorf(errorf, d, t.GetName())
		}

		seen[d] = true
	}

	v.Elem().FieldByName("Dependencies").Set(reflect.ValueOf(deps))
	return nil
}

// ProcesPaths takes paths relative to the target and absolutes them,
// unless they are going to be exported in to the target folder from a dependency.
func (pp *PostProcessor) ProcessPaths(t build.Target, deps []build.Target) error {

	v := reflect.ValueOf(t)

	r := reflect.TypeOf(t).Elem()

	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)

		tag := f.Tag.Get("build")
		if tag != "path" {
			continue
		}

		n := v.Elem().FieldByName(f.Name)

		isExported := func(s string) bool {
			for _, d := range deps {
				if _, ok := d.Installs()[s]; ok {
					return true
				}
			}
			return false
		}

		switch n.Kind() {
		case reflect.String:
			s := n.Interface().(string)
			if isExported(s) {
				continue
			}
			if s == "" {
				continue
			}
			n.SetString(pp.absPath(s))
		case reflect.Slice:
			switch n.Type().Elem().Kind() {
			case reflect.String:
				strs := n.Convert(reflect.TypeOf([]string{})).Interface().([]string)
				for i, s := range strs {
					if isExported(s) {
						continue
					}
					strs[i] = pp.absPath(s)
				}
				n.Set(reflect.ValueOf(strs))
			}
		}

	}

	return nil
}

func (pp *PostProcessor) absPath(s string) string {

	if len(s) < 2 {
		log.Fatalf("%s is invalid", s)
	}
	switch {
	case s[:2] == "//":
		return filepath.Join(pp.projectPath, strings.Trim(s, "//"))
	default:
		if filepath.IsAbs(s) {
			return s
		}
		return filepath.Join(pp.projectPath, pp.packagePath, s)
	}
}

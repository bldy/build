// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package builder parses build graphs and coordinates builds
package postprocessor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"bldy.build/build"
	"bldy.build/build/project"
)

type PostProcessor struct {
	projectPath, packagePath string
}

// New returns a new PostProcessor
func New(p string) PostProcessor {
	return PostProcessor{
		packagePath: p,
		projectPath: project.Root(),
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
			return fmt.Errorf("post process dependencies: %s is duplicated", d)
		} else {
			seen[d] = true
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
		if !(tag == "path" || tag == "expand") {
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
		exp := func(s string) string {
			if tag == "path" {
				return pp.absPath(s)
			}
			if tag == "expand" {
				return os.Expand(s, project.Getenv)
			}
			return s
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
			n.SetString(exp(s))
		case reflect.Slice:
			switch n.Type().Elem().Kind() {
			case reflect.String:
				strs := n.Convert(reflect.TypeOf([]string{})).Interface().([]string)
				for i, s := range strs {
					if isExported(s) {
						continue
					}
					strs[i] = exp(s)
				}
				n.Set(reflect.ValueOf(strs))
			}
		case reflect.Map:
			switch n.Type().Elem().Kind() {
			case reflect.String:
				strs := n.Convert(reflect.TypeOf(map[string]string{})).Interface().(map[string]string)
				for k, v := range strs {
					if isExported(k) {
						continue
					}
					delete(strs, k)

					k, v = exp(k), exp(v)

					strs[k] = v
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
	var r string
	switch {
	case s[:2] == "//":
		r = filepath.Join(pp.projectPath, strings.Trim(s, "//"))
	default:
		r = os.Expand(s, project.Getenv)
		if filepath.IsAbs(r) {
			return r
		}
		r = filepath.Join(pp.projectPath, pp.packagePath, s)
	}
	r = os.Expand(r, project.Getenv)
	return r
}

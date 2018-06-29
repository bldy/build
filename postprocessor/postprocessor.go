// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package builder parses build graphs and coordinates builds
package postprocessor

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"bldy.build/build"
	"bldy.build/build/label"
	"bldy.build/build/project"
	"bldy.build/build/workspace"
)

type PostProcessor struct {
	packagePath string
	ws          workspace.Workspace
}

// New returns a new PostProcessor
func New(ws workspace.Workspace, l label.Label) PostProcessor {
	pkg := ws.AbsPath()
	if l.Package != nil {
		pkg = path.Join(pkg, *l.Package)
	}
	return PostProcessor{
		packagePath: pkg,
		ws:          ws,
	}
}

// ProcessDependencies takes relative dependency paths and turns then in to
// absolute paths.
func (pp *PostProcessor) ProcessDependencies(t build.Rule) error {

	v := reflect.ValueOf(t)

	deps := v.Elem().FieldByName("Dependencies").Interface().([]label.Label)

	seen := make(map[string]bool)

	for _, d := range deps {
		if _, ok := seen[d.String()]; ok {
			return fmt.Errorf("post process dependencies: %s is duplicated", d)
		}
		seen[d.String()] = true

	}

	v.Elem().FieldByName("Dependencies").Set(reflect.ValueOf(deps))
	return nil
}

// ProcessPaths takes paths relative to the target and absolutes them,
// unless they are going to be exported in to the target folder from a dependency.
func (pp *PostProcessor) ProcessPaths(t build.Rule, deps []build.Rule) error {
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
		r = filepath.Join(pp.ws.AbsPath(), strings.Trim(s, "//"))
	default:
		r = os.Expand(s, project.Getenv)
		if filepath.IsAbs(r) {
			return r
		}
		r = filepath.Join(pp.ws.AbsPath(), pp.packagePath, s)
	}
	r = os.Expand(r, project.Getenv)
	return r
}

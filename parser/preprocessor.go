// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser // import "sevki.org/build/parser"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"

	"log"

	"os"
	"os/exec"

	"regexp"

	"strings"

	"sevki.org/build"
	"sevki.org/build/ast"
	"sevki.org/build/util"
)

type PreProcessor struct {
	wd       string
	targs    map[string]build.Target
	document *ast.File
}

func (pp *PreProcessor) Process(d *ast.File) map[string]build.Target {
	pp.targs = make(map[string]build.Target)
	if d == nil {
		log.Fatal("should not be null")
	}
	pp.wd = d.Path
	pp.document = d
	for _, f := range d.Funcs {
		pp.runFunc(f)
	}

	return pp.targs
}
func (pp *PreProcessor) runFunc(f *ast.Func) {
	switch f.Name {
	case "include_defs":
		// takes only one parameter
		packagePath := ""
		switch f.AnonParams[0].(type) {
		case string:
			packagePath = f.AnonParams[0].(string)
			break
		default:
			log.Fatal("include_defs takes a string as an argument")
		}

		document, err := ReadBuildFile(TargetURL{Package: packagePath}, pp.wd)
		if err != nil {
			log.Fatal(err)
		}
		if pp.document.Vars == nil {
			pp.document.Vars = make(map[string]interface{})
		}
		for k, v := range document.Vars {
			pp.document.Vars[k] = v
		}
	case "load":

		filePath := ""
		var varsToImport []string
		// Check paramter types
		for i, param := range f.AnonParams {
			switch param.(type) {
			case string:
				if i == 0 {
					filePath = param.(string)
				} else {
					varsToImport = append(varsToImport, param.(string))
				}
				break
			default:
				log.Fatal("should be used like so; load(file, var...)")
			}
		}

		document, err := ReadFile(pp.absPath(filePath))

		if err != nil {
			log.Fatal(err)
		}
		if pp.document.Vars == nil {
			pp.document.Vars = make(map[string]interface{})
		}

		for _, v := range varsToImport {
			if val, ok := document.Vars[v]; ok {
				pp.document.Vars[v] = val
			} else {
				log.Fatalf("%s is not present at %s. Please check the file and try again.", v, filePath)
			}
		}
	case "select":
	default:
		targ, err := pp.makeTarget(f)
		if err != nil {
			log.Fatal(err)
			return
		}
		pp.targs[targ.GetName()] = targ
	}
}

func (pp *PreProcessor) absPath(s string) string {
	if strings.TrimLeft(s, "//") != s {
		return filepath.Join(util.GetProjectPath(), strings.Trim(s, "//"))
	} else {
		return filepath.Join(pp.document.Path, s)
	}
}

func (pp *PreProcessor) makeTarget(f *ast.Func) (build.Target, error) {

	ttype := ast.Get(f.Name)

	payload := make(map[string]interface{})

	for key, fn := range f.Params {

		field, err := ast.GetFieldByTag(f.Name, key, ttype)
		if err != nil {
			return nil, err
		}

		var i interface{}
		switch fn.(type) {
		case *ast.Func:
			x := fn.(*ast.Func)
			i = pp.funcReturns(x)
		case ast.Variable:
			i = pp.document.Vars[fn.(ast.Variable).Value]
		default:
			i = fn
		}

		if field.Type != reflect.TypeOf(i) {
			// return nil, fmt.Errorf("%s is of type %s not %s.", key, reflect.TypeOf(i).String(), field.Type.String())
		}

		btag := field.Tag.Get("build")

		if field.Name == "Dependencies" {
			var deps []string
			for _, d := range i.([]interface{}) {
				ds := d.(string)

				switch {
				case ds[:2] == "//":
					deps = append(deps, ds)
					break
				case ds[0] == ':':
					r, _ := filepath.Rel(util.GetProjectPath(), pp.wd)
					deps = append(deps,
						fmt.Sprintf("//%s%s", r, ds),
					)
					break
				default:
					errorf := `dependency '%s' in %s is not a valid URL for a target.
a target url can only start with a '//' or a ':' for relative targets.`

					log.Fatalf(errorf, ds, f.Params["name"])
				}
			}
			i = deps
			goto SKIP
		}

		if btag == "path" {
			var files []string
			switch i.(type) {
			case []interface{}:
				for _, s := range i.([]interface{}) {
					absPath, _ := s.(string)
					files = append(files, absPath)
				}
			case []string:
				files = i.([]string)
			case string:
				i = pp.absPath(i.(string))
				goto SKIP
			}

			for n, x := range files {
				files[n] = pp.absPath(x)
			}
			i = files
		}
	SKIP:
		payload[field.Name] = i

	}

	//BUG(sevki): this is a very hacky way of doing this but it seems to be safer don't mind.
	var bytz []byte
	buf := bytes.NewBuffer(bytz)

	enc := json.NewEncoder(buf)
	enc.Encode(payload)

	t := reflect.New(ttype).Interface()
	dec := json.NewDecoder(buf)
	dec.Decode(t)
	switch t.(type) {
	case build.Target:
		break
	default:
		log.Fatalf("type %s doesn't implement the build.Target interface, check sevki.co/2LLRfc for more information", ttype.String())
	}
	return t.(build.Target), nil
}

func (pp *PreProcessor) funcReturns(f *ast.Func) interface{} {
	switch f.Name {
	case "glob":
		return pp.glob(f)
	case "version":
		return pp.version(f)
	}
	return ""
}

func (pp *PreProcessor) glob(f *ast.Func) []string {
	if !filepath.IsAbs(pp.wd) {
		return []string{fmt.Sprintf("Error parsing glob: %s is not an absolute path.", pp.wd)}
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
			globPtrn = filepath.Join(pp.wd, s.(string))
		default:
			return nil
		}

		globFiles, err := filepath.Glob(globPtrn)

		if err != nil {
			return []string{"Error parsing glob: %s"}
		}
	DONE:
		for i, f := range globFiles {

			t, _ := filepath.Rel(pp.document.Path, f)

			for _, x := range excludes {
				if x.Match([]byte(t)) {
					globFiles = append(globFiles[:i], globFiles[i+1:]...)
					goto DONE
				}
			}
			globFiles[i] = t

		}
		files = append(files, globFiles...)
	}
	return files
}

func (pp *PreProcessor) env(f *ast.Func) string {
	if len(f.AnonParams) != 1 {
		return ""
	}
	return os.Getenv(f.AnonParams[0].(string))
}

func (pp *PreProcessor) version(f *ast.Func) string {

	if out, err := exec.Command("git",
		"--git-dir="+util.GetGitDir(pp.wd)+".git",
		"describe",
		"--always").Output(); err != nil {
		return err.Error()
	} else {
		return strings.TrimSpace(string(out))
	}
}

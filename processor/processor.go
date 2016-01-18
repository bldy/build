// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package processor // import "sevki.org/build/processor"

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
	"sevki.org/build/parser"
	"sevki.org/build/util"
)

type Processor struct {
	wd       string
	targs    map[string]build.Target
	document *ast.File
}

func (p *Processor) Process(d *ast.File) map[string]build.Target {
	p.targs = make(map[string]build.Target)
	if d == nil {
		log.Fatal("should not be null")
	}
	p.wd = d.Path
	p.document = d

	for k, v := range d.Vars {
		switch v.(type) {
		case *ast.Func:
			d.Vars[k] = p.funcReturns(v.(*ast.Func))
		}
	}

	for _, f := range d.Funcs {
		p.runFunc(f)
	}

	return p.targs
}
func (p *Processor) runFunc(f *ast.Func) {
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

		document, err := parser.ReadBuildFile(parser.TargetURL{Package: packagePath}, p.wd)
		if err != nil {
			log.Fatal(err)
		}
		if p.document.Vars == nil {
			p.document.Vars = make(map[string]interface{})
		}
		for k, v := range document.Vars {
			p.document.Vars[k] = v
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

		document, err := parser.ReadFile(p.absPath(filePath))

		if err != nil {
			log.Fatal(err)
		}
		if p.document.Vars == nil {
			p.document.Vars = make(map[string]interface{})
		}

		for _, v := range varsToImport {

			if val, ok := document.Vars[v]; ok {
				p.document.Vars[v] = val
			} else {
				log.Fatalf("%s is not present at %s. Please check the file and try again.", v, filePath)
			}
		}
	case "select":
	default:
		targ, err := p.makeTarget(f)
		if err != nil {
			log.Fatal(err)
			return
		}
		p.targs[targ.GetName()] = targ
	}
}

func (p *Processor) absPath(s string) string {
	if strings.TrimLeft(s, "//") != s {
		return filepath.Join(util.GetProjectPath(), strings.Trim(s, "//"))
	} else {
		return filepath.Join(p.document.Path, s)
	}
}

func (p *Processor) makeTarget(f *ast.Func) (build.Target, error) {

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
			i = p.funcReturns(x)
		case ast.Variable:
			i = p.document.Vars[fn.(ast.Variable).Key]
		default:
			i = fn
		}

		if field.Type != reflect.TypeOf(i) {
			// return nil, fmt.Errorf("%s is of type %s not %s.", key, reflect.TypeOf(i).String(), field.Type.String())
		}

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

func (p *Processor) funcReturns(f *ast.Func) interface{} {
	switch f.Name {
	case "glob":
		return p.glob(f)
	case "version":
		return p.version(f)
	case "addition":
		return p.combineArrays(f)
	}
	return ""
}

func (p *Processor) combineArrays(f *ast.Func) interface{} {
	var t []interface{}

	for _, v := range f.AnonParams {
		switch v.(type) {
		case ast.Variable:
			name := v.(ast.Variable)
			x := p.document.Vars[name.Key]
			switch x.(type) {
			case []interface{}:
				t = append(t, x.([]interface{})...)
			}
		}
	}

	return t
}

func (p *Processor) glob(f *ast.Func) []string {
	if !filepath.IsAbs(p.wd) {
		return []string{fmt.Sprintf("Error parsing glob: %s is not an absolute path.", p.wd)}
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
			globPtrn = filepath.Clean(filepath.Join(p.wd, s.(string)))
			log.Println(globPtrn)
		default:
			return nil
		}

		globFiles, err := filepath.Glob(globPtrn)

		if err != nil {
			return []string{"Error parsing glob: %s"}
		}
	RESIZED:
		for i, f := range globFiles {
			t, _ := filepath.Rel(util.GetProjectPath(), f)
			t = fmt.Sprintf("//%s", t)
			for _, x := range excludes {
				if x.Match([]byte(t)) {
					globFiles = append(globFiles[:i], globFiles[i+1:]...)
					goto RESIZED
				}
			}
			globFiles[i] = t

		}
		files = append(files, globFiles...)
	}
	return files
}

func (p *Processor) env(f *ast.Func) string {
	if len(f.AnonParams) != 1 {
		return ""
	}
	return os.Getenv(f.AnonParams[0].(string))
}

func (p *Processor) version(f *ast.Func) string {

	if out, err := exec.Command("git",
		"--git-dir="+util.GetGitDir(p.wd)+".git",
		"describe",
		"--always").Output(); err != nil {
		return err.Error()
	} else {
		return strings.TrimSpace(string(out))
	}
}

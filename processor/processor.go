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
	"sevki.org/build/internal"
	"sevki.org/build/parser"
	"sevki.org/build/token"
	"sevki.org/build/util"
)

type Processor struct {
	vars    map[string]interface{}
	wd      string
	targs   map[string]build.Target
	parser  *parser.Parser
	Targets chan *build.Target
}

func NewProcessor(p *parser.Parser) *Processor {
	return &Processor{
		vars:    make(map[string]interface{}),
		parser:  p,
		Targets: make(chan *build.Target),
	}
}
func (p *Processor) Run() {

	go p.parser.Run()
	var d ast.Decl
	d = <-p.parser.Decls
	for ; d != nil; d = <-p.parser.Decls {
		switch d.(type) {
		case *ast.Func:
			p.runFunc(d.(*ast.Func))
		case *ast.Assignment:
			p.doAssignment(d.(*ast.Assignment))
		}
	}
	p.Targets <- nil

}

func (p *Processor) doAssignment(a *ast.Assignment) {
 	switch a.Value.(type) {
		case ast.BasicLit:
		p.vars[a.Key] = a.Value.(ast.BasicLit).Interface()
		case *ast.Func:
		case ast.Slice:
		p.vars[a.Key] = a.Value.(ast.Slice).Slice
		default:
		log.Printf("%T", a.Value)
	}
}
 
func (p *Processor) runFunc(f *ast.Func) {
	switch f.Name {

	case "load":
		fail := func() {
			log.Fatal("should be used like so; load(file, var...)")
		}

		filePath := ""
		var varsToImport []string
		// Check paramter types
		for i, param := range f.AnonParams {
			switch param.(type) {
			case ast.BasicLit:
				v := param.(ast.BasicLit)
				if v.Kind != token.Quote {
					fail()
				}
				if i == 0 {
					filePath = v.Value
				} else {
					varsToImport = append(varsToImport, v.Value)
				}
				break
			default:
				fail()
			}
		}

		loadingProcessor, err := NewProcessorFromFile(p.absPath(filePath))
		if err != nil {
			log.Fatal(err)
		}
		if p.vars == nil {
			p.vars = make(map[string]interface{})
		}

		for _, v := range varsToImport {

			if val, ok := loadingProcessor.vars[v]; ok {
				p.vars[v] = val
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
		return filepath.Join(p.parser.Path, s)
	}
}

func NewProcessorFromFile(n string) (*Processor, error) {

	ks, err := os.Open(n)
	if err != nil {
		return nil, fmt.Errorf("opening file: %s\n", err.Error())
	}
	ts, _ := filepath.Abs(ks.Name())
	dir := strings.Split(ts, "/")
	p := parser.New("BUILD", "/"+filepath.Join(dir[:len(dir)-1]...), ks)

	return NewProcessor(p), nil
}

func (p *Processor) makeTarget(f *ast.Func) (build.Target, error) {
	ttype := internal.Get(f.Name)

	payload := make(map[string]interface{})

	for key, fn := range f.Params {

		field, err := internal.GetFieldByTag(f.Name, key, ttype)
		if err != nil {
			return nil, err
		}

		var i interface{}
		switch fn.(type) {
		case *ast.Func:
			x := fn.(*ast.Func)
			i = p.funcReturns(x)
		case ast.Variable:
			i = p.vars[fn.(ast.Variable).Key]
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
			x, exists := p.vars[name.Key]
			if !exists {
				log.Fatalf("combinine arrays: coudln't find var %s", name.Key)
			}
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

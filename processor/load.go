// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package processor // import "bldy.build/build/processor"
import "bldy.build/build/ast"

var (
	cache = make(map[string]map[string]interface{})
)

func (p *Processor) load(f *ast.Func) {
	fail := func() {
		p.l.Fatal("should be used like so; load(file, var...)")
	}

	filePath := ""
	var varsToImport []string
	// Check paramter types
	for i, param := range f.AnonParams {
		switch param.(type) {
		case string:
			v := param.(string)
			if i == 0 {
				filePath = v
			} else {
				varsToImport = append(varsToImport, v)
			}
			break
		default:
			fail()
		}
	}
	file := p.absPath(filePath)

	if _, ok := cache[file]; !ok {
		loadingProcessor, err := NewProcessorFromFile(file)
		if err != nil {
			p.l.Fatal(err)
		}
		go loadingProcessor.Run()
		for d := <-loadingProcessor.Targets; d != nil; d = <-loadingProcessor.Targets {
		}
		cache[file] = loadingProcessor.vars
	}

	if p.vars == nil {
		p.vars = make(map[string]interface{})
	}

	for _, v := range varsToImport {
		if val, ok := cache[file][v]; ok {
			p.vars[v] = val
		} else {
			p.l.Fatalf("%s is not present at %s. Please check the file and try again.", v, filePath)
		}
	}
}

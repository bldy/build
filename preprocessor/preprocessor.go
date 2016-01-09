// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package preprocessor // import "sevki.org/build/preprocessor"
import (
	"fmt"
	"log"
	"path/filepath"

	"sevki.org/build/ast"
)

func Process(f *ast.File) error {
	return processDupeLoad(f)
}

func processDupeLoad(f *ast.File) error {
	seenFile := make(map[string]*ast.Func)
	for _, function := range f.Funcs {
		if function.Name != "load" {
			continue
		}
		var fileName string
		switch function.AnonParams[0].(type) {
		case string:
			fileName = function.AnonParams[0].(string)
		default:
			errorMessage := `load must always be in this form 'load("//foo/bar/FILE", "EXPORTED_VALUE_A", "EXPORTED_VALUE_B")'`
			log.Fatal(errorMessage)
		}

		if before, seen := seenFile[fileName]; seen {
			return fmt.Errorf("'load' function in file %s, loads from same file %s twice. try merging load functions on line %d and %d.",
				filepath.Join(f.Path, function.File),
				function.AnonParams[0].(string),
				function.Line,
				before.Line,
			)
		} else {
			seenFile[fileName] = function
		}
	}
	return nil
}

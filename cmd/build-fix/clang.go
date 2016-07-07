package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/go-clang/v3.7/clang"
)

type fix struct {
	startLine, startCol, endLine, endCol uint16
	str                                  string
}

func doe(err error) {
	if err != nil {
		log.Fatal(err)
	}

}

func fixit(src string, flags []string, w bool) {
	if !filepath.IsAbs(src) {
		return
	}
	idx := clang.NewIndex(0, 1)
	defer idx.Dispose()

	tu := idx.ParseTranslationUnit(src, flags, nil, 0)
	defer tu.Dispose()

	var fixs []fix
	for _, diag := range tu.Diagnostics() {

		fixit, str := diag.FixIt(0)
		_, startLine, startCol, _ := fixit.Start().ExpansionLocation()
		_, endLine, endCol, _ := fixit.End().ExpansionLocation()

		fixs = append(fixs, fix{
			startLine: startLine,
			startCol:  startCol,
			endLine:   endLine,
			endCol:    endCol,
			str:       str,
		})
	}

	f, _ := os.Open(src)
	scanner := bufio.NewScanner(f)
	var line uint16 = 0
	out := os.Stdout
	if w {
		var err error
		out, err = ioutil.TempFile("", "")
		doe(err)

		defer func() {
			doe(out.Close())
			doe(f.Close())
			doe(os.Rename(out.Name(), src))
		}()
	}
	for scanner.Scan() {
		line++
		lineBuf := scanner.Text()
		fixed := false
		for _, fix := range fixs {
			if fix.startLine == line {
				fmt.Fprint(out, lineBuf[:fix.startCol-1])
				fmt.Fprint(out, fix.str)
				for {
					if line != fix.endLine {
						lineBuf = scanner.Text()
						line++
					}
					fmt.Fprintln(out, lineBuf[fix.endCol-1:])
					break
				}
				fixed = true
			}
		}

		if !fixed {
			fmt.Fprintln(out, scanner.Text())
		}

	}
}

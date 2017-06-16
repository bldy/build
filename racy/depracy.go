// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package racy // import "bldy.build/build/racy"
import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"bldy.build/build/project"
)

var fc = make(chan string)

// HashFiles will hash files collecetion represented as a string array,
// If the string in the array is directory it will the directory contents to the array
// if the string isn't an absolute path, it will assume that it's a export from a dependency
// and skip that.
func HashFiles(h io.Writer, files []string) {
	for _, fyl := range files {
		if !filepath.IsAbs(fyl) {
			continue
		}
	}
	fsm := files
RESTART:
	for i, file := range fsm {
		if !filepath.IsAbs(file) {
			continue
		}
		if filepath.Base(file) == project.BuildOut() {
			continue
		}

		f, err := os.Open(file)

		if err != nil {
			log.Fatalf("hash files: %s\n", err.Error())
		}

		stat, _ := f.Stat()
		if stat.IsDir() {
			fsm = append([]string{}, fsm[i+1:]...)
			fs, _ := f.Readdir(-1)
			for _, x := range fs {
				fsm = append(fsm, (filepath.Join(file, x.Name())))
			}
			f.Close()
			goto RESTART /* to avoid out of bound errors, there may be no files
			in the folder */
		}

		fmt.Fprintf(h, "file %s\n", filepath.Join(project.Root(), file))
		n, _ := io.Copy(h, f)
		fmt.Fprintf(h, "%d bytes\n", n)
		f.Close()
	}
}

// HashFilesWithExt will hash files collecetion represented as a string array,
// If the string in the array is directory it will the directory contents to the array
// if the string isn't an absolute path, it will assume that it's a export from a dependency
// and skip that.
func HashFilesWithExt(h io.Writer, files []string, ext string) {
	for _, fyl := range files {
		if !filepath.IsAbs(fyl) {
			continue
		}
		fc <- fyl
	}
	fsm := files
RESTART:
	for i, file := range fsm {
		if !filepath.IsAbs(file) {
			continue
		}
		if filepath.Base(file) == project.BuildOut() {
			continue
		}
		f, err := os.Open(file)

		if err != nil {
			log.Fatalf("hash files: %s\n", err.Error())
		}

		stat, _ := f.Stat()
		if stat.IsDir() {
			fsm = append([]string{}, fsm[i+1:]...)
			fs, _ := f.Readdir(-1)
			for _, x := range fs {
				if filepath.Ext(x.Name()) == ext || filepath.Ext(x.Name()) == "" {
					fsm = append(fsm, (filepath.Join(file, x.Name())))
				}

			}
			goto RESTART /* to avoid out of bound errors, there may be no files
			in the folder */
		}
		if filepath.Ext(file) != ext {
			f.Close()
			continue
		}

		fmt.Fprintf(h, "file %s\n", filepath.Join(project.Root(), file))
		n, _ := io.Copy(h, f)
		fmt.Fprintf(h, "%d bytes\n", n)
		f.Close()
	}
}

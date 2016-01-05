// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	pp = ""
)

func init() {
	wd, _ := os.Getwd()
	pp = GetGitDir(wd)
}
func GetProjectPath() (ProjectPath string) {
	return pp
}
func RelPPath(p string) string {
	rel, _ := filepath.Rel(GetProjectPath(), p)
	return rel
}

func HashFiles(h io.Writer, files []string) {
	fsm := files
RESTART:
	for i, file := range fsm {
		f, err := os.Open(file)

		if err != nil {
			log.Fatalf("%s error\n", filepath.Join(pp, file))
		}
		stat, _ := f.Stat()
		if stat.IsDir() {
			fsm = append([]string{}, fsm[i+1:]...)
			fs, _ := f.Readdir(-1)
			for _, x := range fs {
				fsm = append(fsm, (filepath.Join(file, x.Name())))
			}
			goto RESTART /* to avoid out of bound errors, there may be no files
			in the folder */
		}

		fmt.Fprintf(h, "file %s\n", filepath.Join(pp, file))
		n, _ := io.Copy(h, f)
		fmt.Fprintf(h, "%d bytes\n", n)
		f.Close()
	}
}

func HashStrings(h io.Writer, strs []string) {
	for _, str := range strs {
		io.WriteString(h, str)
	}
}
func GetGitDir(p string) string {
	dirs := strings.Split(p, "/")
	for i := len(dirs) - 1; i > 0; i-- {
		try := fmt.Sprintf("/%s/.git", filepath.Join(dirs[0:i+1]...))
		if _, err := os.Lstat(try); os.IsNotExist(err) {
			continue
		} else if err != nil {
			log.Fatal(err)
		}
		pr, _ := filepath.Split(try)
		return pr
	}
	return ""
}

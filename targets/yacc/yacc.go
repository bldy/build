// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "sevki.org/build/targets/yacc"

import (
	"bytes"
	"crypto/sha1"

	"io"
	"os/exec"
	"path/filepath"

	"log"

	"strings"

	"fmt"

	"sevki.org/build/ast"
	"sevki.org/build/build"
	"sevki.org/build/util"
)

var YaccVersion = ""

type Yacc struct {
	Name         string   `yacc:"name"`
	Sources      []string `yacc:"srcs" build:"path"`
	Dir          string   `yacc:"dir" build:"path"`
	Exports      []string `yacc:"exports"`
	Dependencies []string `yacc:"deps"`
	YaccOptions  []string `yacc:"yaccopts"`
	Source       string
	buf          bytes.Buffer
}

func init() {
	if out, err := exec.Command("yacc", "--version").Output(); err != nil {
		YaccVersion = "deadbeef"
	} else {
		YaccVersion = strings.TrimSpace(string(out))
	}

	if err := ast.Register("yacc", Yacc{}); err != nil {
		log.Fatal(err)
	}
}
func (y *Yacc) Hash() []byte {
	h := sha1.New()
	io.WriteString(h, YaccVersion)
	io.WriteString(h, y.Name)
	util.HashFiles(h, y.Sources)
	util.HashStrings(h, y.YaccOptions)
	return h.Sum(nil)
}

func (y *Yacc) Build(c *build.Context) error {

	params := []string{}
	params = append(params, y.YaccOptions...)
	params = append(params, y.Sources...)

	c.Println(strings.Join(append([]string{"yacc"}, params...), " "))

	if err := c.Exec("yacc", nil, params); err != nil {
		c.Println(err.Error())
		return fmt.Errorf(y.buf.String())
	}

	for _, export := range y.Exports {
		c.Println(strings.Join([]string{"cp", export, filepath.Join(y.Dir, export)}, " "))
		if err := c.Exec("cp", nil, []string{export, filepath.Join(y.Dir, export)}); err != nil {
			c.Println(err.Error())
			return fmt.Errorf(y.buf.String())
		}
	}
	return nil
}
func (y *Yacc) Installs() map[string]string {
	return nil
}
func (y *Yacc) GetName() string {
	return y.Name
}

func (y *Yacc) GetDependencies() []string {
	return y.Dependencies
}
func (y *Yacc) GetSource() string {
	return y.Source
}

func (y *Yacc) Reader() io.Reader {
	return &y.buf
}

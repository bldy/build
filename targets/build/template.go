// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build // import "bldy.build/build/targets/build"

import (
	"crypto/sha1"
	"html/template"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"sevki.org/lib/prettyprint"

	"bldy.build/build"
	"bldy.build/build/racy"
	"bldy.build/build/targets/cc"
)

type Template struct {
	Name         string                 `template:"name"`
	Dependencies []string               `template:"deps"`
	Template     string                 `template:"template" build:"path"`
	Out          string                 `template:"out"`
	Vars         map[string]interface{} `template:"vars"`
}

func (t *Template) GetName() string {
	return t.Name
}

func (t *Template) GetDependencies() []string {
	return nil
}

func (t *Template) Hash() []byte {
	t.Vars["_CCVER"] = strings.Split(cc.CCVersion, "\n")[0]

	h := sha1.New()
	io.WriteString(h, prettyprint.AsJSON(t))
	return racy.XOR(h.Sum(nil),
		racy.HashFilesForExt([]string{t.Template}, filepath.Ext(t.Template)))
}

func (t *Template) Build(c *build.Runner) error {
	t.Vars["_CCVER"] = strings.Split(cc.CCVersion, "\n")[0]

	_, file := filepath.Split(t.Template)
	bytz, err := ioutil.ReadFile(t.Template)
	if err != nil {
		return err
	}
	tmpl, err := template.New(file).Parse(string(bytz))
	if err != nil {
		return err
	}
	_, ofile := filepath.Split(t.Out)

	outfile, err := c.Create(ofile)
	if err != nil {
		return err
	}
	err = tmpl.Execute(outfile, t.Vars)
	if err != nil {
		return err
	}
	return nil
}

func (t *Template) Installs() map[string]string {
	installs := make(map[string]string)
	_, file := filepath.Split(t.Out)
	installs[t.Out] = file
	return installs
}

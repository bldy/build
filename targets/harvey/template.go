// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harvey // import "bldy.build/build/targets/harvey"
import (
	"html/template"
	"io/ioutil"
	"path/filepath"

	"bldy.build/build"
	"bldy.build/build/racy"
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
	return racy.HashFilesForExt([]string{t.Template}, filepath.Ext(t.Template))
}

func (t *Template) Build(c *build.Context) error {
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

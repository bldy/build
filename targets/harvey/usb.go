// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harvey

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"text/template"

	"bldy.build/build"
	"bldy.build/build/racy"
)

type Embed struct {
	Name  string
	Class string
	CSP   [4]string
	VID   string
	DID   string
}

type SysConf struct {
	Embeds []Embed
}

type USB struct {
	Name         string   `usb:"name"`
	Dependencies []string `usb:"deps"`
	Conf         string   `usb:"conf" build:"path"`
}

func (u *USB) GetName() string {
	return u.Name
}

func (u *USB) GetDependencies() []string {
	return u.Dependencies
}

func (u *USB) Hash() []byte {
	h := sha1.New()
	racy.HashFiles(h, []string{u.Conf})
	io.WriteString(h, u.Name)
	return []byte{}
}
func (s *USB) Installs() map[string]string {
	installs := make(map[string]string)
	fileOut := fmt.Sprintf("%s.c", s.Name)
	installs[fileOut] = fileOut
	return installs
}

func (s *USB) Build(c *build.Context) error {
	f, err := c.Open(s.Conf)
	if err != nil {
		return err
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	var sysconf SysConf
	err = json.Unmarshal(buf, &sysconf)
	if err != nil {
		return err
	}
	fileOut := fmt.Sprintf("%s.c", s.Name)
	out, err := c.Create(fileOut)
	if err != nil {
		return err
	}
	tmpl, err := template.New("usbdb.c").Parse(jtab)
	err = tmpl.Execute(out, sysconf.Embeds)
	if err != nil {
		return err
	}
	return nil
}

var jtab = `
/* machine generated. do not edit */
#include <u.h>
#include <libc.h>
#include <thread.h>
#include <usb/usb.h>
#include "usbd.h"

{{ range . }}	int {{ .Name }}main(Dev*, int, char**);
{{ end }}

Devtab devtab[] = {
	/* device, entrypoint, {csp, csp, csp csp}, vid, did */
{{ range . }}	{ "{{ .Name}}", {{ .Name }}main,  { {{ .Class}} | {{ index .CSP 0}},{{index .CSP 1}},{{ index .CSP 2}}, {{ index .CSP 3}}  }, {{ .VID}}, {{.DID}}, ""},
{{ end }}
	{nil, nil,	{0, 0, 0, 0, }, -1, -1, nil},
};

/* end of machine generated */

`

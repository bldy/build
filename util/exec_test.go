// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestPass(t *testing.T) {
	var stdOut, stdErr bytes.Buffer
	if err := Exec(
		&stdOut,
		&stdErr,
		"go",
		nil,
		[]string{"version"},
	); err != nil {
		t.Error(err)
	}
}

func TestFail(t *testing.T) {
	var stdOut, stdErr bytes.Buffer
	if err := Exec(
		&stdOut,
		&stdErr,
		"go",
		nil,
		[]string{"--version"},
	); err == nil {
		t.Fail()
	}
}

func TestCC(t *testing.T) {
	if err := os.Chdir("/tmp"); err != nil {
		t.Error(err)
	}
	helloWorldFile := "helloworld.c"
	hello, err := os.Create(helloWorldFile)
	if err != nil {
		t.Error(err)
	}

	knr := `main( ) {
        printf("hello, world");
}
`

	io.WriteString(hello, knr)
	var stdOut, stdErr bytes.Buffer

	if err := Exec(
		&stdOut,
		&stdErr,
		"cc",
		nil,
		[]string{
			"-o",
			"helloworld",
			"-ansi",
			helloWorldFile,
		},
	); err != nil {
		io.Copy(os.Stderr, &stdErr)
		t.Error(err)
		return
	}

	os.Remove(helloWorldFile)
	os.Remove("helloworld")
}

// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package racy

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"bldy.build/build"
	"bldy.build/build/executor"
)

// regular target
type testTarget struct {
	Name         string
	Dependencies []string
	Srcs         []string
}

func (t *testTarget) GetName() string { return "" }

func (t *testTarget) GetDependencies() []string { return nil }

func (t *testTarget) Hash() []byte { return nil }

func (t *testTarget) Build(*executor.Executor) error { return nil }

func (t *testTarget) Installs() map[string]string { return nil }

// target with path
type testTargetWithPath struct {
	Name         string
	Dependencies []string
	Srcs         []string `build:"path"`
}

func (t *testTargetWithPath) GetName() string { return "" }

func (t *testTargetWithPath) GetDependencies() []string { return nil }

func (t *testTargetWithPath) Hash() []byte { return nil }

func (t *testTargetWithPath) Build(*executor.Executor) error { return nil }

func (t *testTargetWithPath) Installs() map[string]string { return nil }

// With path and extensions
type testTargetWithPathAndExt struct {
	Name         string
	Dependencies []string
	Srcs         []string `build:"path" ext:".c"`
}

func (t *testTargetWithPathAndExt) GetName() string { return "" }

func (t *testTargetWithPathAndExt) GetDependencies() []string { return nil }

func (t *testTargetWithPathAndExt) Hash() []byte { return nil }

func (t *testTargetWithPathAndExt) Build(*executor.Executor) error { return nil }

func (t *testTargetWithPathAndExt) Installs() map[string]string { return nil }

func TestHashTarget(t *testing.T) {
	testTarg := &testTarget{
		Name: "foo",
	}

	if bytes.Compare(HashTarget(testTarg), []byte{}) == 0 {
		t.Fail()
	}

}

func TestHashesForDifferent(t *testing.T) {
	dir, _ := ioutil.TempDir("", "racy_tests")
	defer os.RemoveAll(dir)
	for name, content := range map[string]string{
		"hello.c": "#include bla",
		"hello.h": "#define __HELLO_",
	} {
		tmpfn := filepath.Join(dir, name)

		if err := ioutil.WriteFile(tmpfn, []byte(content), 0666); err != nil {
			log.Fatal(err)
		}

	}

	tests := []struct {
		name string
		a, b build.Target
		comp int
	}{
		{
			name: "names",
			a: &testTarget{
				Name: "foo",
			},
			b: &testTarget{
				Name: "boo",
			},
			comp: 0,
		},
		{
			name: "dependencies",
			a: &testTarget{
				Name:         "foo",
				Dependencies: []string{"bar"},
			},
			b: &testTarget{
				Name:         "foo",
				Dependencies: []string{"bar", "bar"},
			},
			comp: 0,
		},
		{
			name: "files",
			a: &testTarget{
				Name: "foo",
				Srcs: []string{dir},
			},
			b: &testTargetWithPath{
				Name: "foo",
				Srcs: []string{dir},
			},
			comp: 0,
		},
		{
			name: "filesWithExts",
			a: &testTargetWithPath{
				Name: "foo",
				Srcs: []string{dir},
			},
			b: &testTargetWithPathAndExt{
				Name: "foo",
				Srcs: []string{dir},
			},
			comp: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hashA := HashTarget(test.a)
			hashB := HashTarget(test.b)
			if bytes.Compare(hashA, hashB) == 0 {
				t.Logf("\"%x\" and \"%x\" are the same", hashA, hashB)
				t.Fail()
			}
		})
	}

}

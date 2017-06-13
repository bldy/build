// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package racy

import (
	"bytes"
	"testing"

	"bldy.build/build"
)

type testTarget struct {
	Name         string
	Dependencies []string
	Srcs         []string
}
type testTargetWithPath struct {
	Name         string
	Dependencies []string
	Srcs         []string `build:"path"`
}

func (t *testTarget) GetName() string { return "" }

func (t *testTarget) GetDependencies() []string { return nil }

func (t *testTarget) Hash() []byte { return nil }

func (t *testTarget) Build(*build.Executor) error { return nil }

func (t *testTarget) Installs() map[string]string { return nil }
func (t *testTargetWithPath) GetName() string     { return "" }

func (t *testTargetWithPath) GetDependencies() []string { return nil }

func (t *testTargetWithPath) Hash() []byte { return nil }

func (t *testTargetWithPath) Build(*build.Executor) error { return nil }

func (t *testTargetWithPath) Installs() map[string]string { return nil }

func TestHashTarget(t *testing.T) {
	testTarg := &testTarget{
		Name: "foo",
	}

	if bytes.Compare(HashTarget(testTarg), []byte{}) == 0 {
		t.Fail()
	}

}

func TestHashesForDifferent(t *testing.T) {
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
			name: "targets",
			a: &testTarget{
				Name: "foo",
				Srcs: []string{"/etc/hosts"},
			},
			b: &testTargetWithPath{
				Name: "foo",
				Srcs: []string{"/etc/hosts"},
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

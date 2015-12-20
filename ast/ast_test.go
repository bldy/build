// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ast defines build data structures.
package ast // import "sevki.org/build/ast"

import "testing"

type TestTarget struct{}
type TestBadTarget struct{}

func (t *TestTarget) Build() error {
	return nil
}

func TestRegister(t *testing.T) {

	if err := Register("test_target", TestTarget{}); err != nil {
		t.Error(err)
	}
}
func TestRegisterBadTarget(t *testing.T) {

	if err := Register("", TestBadTarget{}); err == nil {
		t.Fail()
	}
}

func TestGet(t *testing.T) {

	if err := Register("test_target", TestTarget{}); err != nil {
		t.Error(err)
	}
	if target := Get("test_target"); target == nil {
		t.Error("We couldn't get it")

	}
}

func TestKeyMismatch(t *testing.T) {

	if err := Register("test_target", TestTarget{}); err != nil {
		t.Error(err)
	}
	if target := Get("toast_target"); target != nil {
		t.Error("We couldn't get it")

	}
}

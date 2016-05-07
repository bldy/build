// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package processor // import "sevki.org/build/processor"
import (
	"os"
	"testing"

	"sevki.org/build/targets/cc"
)

func TestSimpleAssignment(t *testing.T) {
	p, err := NewProcessorFromFile("tests/simpleAssignment.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	<-p.Targets
	if p.vars["NAME"] != "build" {
		t.Fail()
	}
}

func TestSliceAssignment(t *testing.T) {
	slc := []string{
		"date.c",
		"time.c",
		"bla.c",
	}

	p, err := NewProcessorFromFile("tests/sliceAssignment.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	<-p.Targets
	for i, v := range p.vars["C_SRCS"].([]interface{}) {
		if slc[i] != v {
			t.Fail()
		}
	}
}

func TestSliceAssignmentWithVariable(t *testing.T) {
	slc := []string{
		"date.c",
		"time.c",
		"bla.c",
	}

	p, err := NewProcessorFromFile("tests/sliceAssignmentWithVariable.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	<-p.Targets
	for i, v := range p.vars["C_SRCS"].([]interface{}) {
		if slc[i] != v {
			t.Fail()
		}
	}
}

func TestMapAssignment(t *testing.T) {
	slc := map[string]string{
		"amd64": "amd64/kmain.c",
		"riscv": "riscv/kmain.com",
	}

	p, err := NewProcessorFromFile("tests/mapAssignment.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	<-p.Targets
	for i, v := range p.vars["C_SRCS"].(map[string]interface{}) {
		if slc[i] != v {
			t.Log(v)
			t.Fail()
		}
	}
}

func TestMapFunc(t *testing.T) {
	slc := map[string]string{
		"GOPATH": os.Getenv("GOPATH"),
		"riscv":  "riscv/kmain.com",
	}

	p, err := NewProcessorFromFile("tests/mapAssignmentFunc.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	<-p.Targets
	for i, v := range p.vars["C_SRCS"].(map[string]interface{}) {
		if slc[i] != v {
			t.Logf("%s is not %s", v, slc[i])
			t.Fail()
		}
	}
}

func TestAddition(t *testing.T) {
	slc := []string{
		"date.c",
		"time.c",
		"bla.c",
	}

	p, err := NewProcessorFromFile("tests/addition.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	<-p.Targets
	for i, v := range p.vars["C_SRCS"].([]interface{}) {
		if slc[i] != v {
			t.Fail()
		}
	}
}

func TestTarget(t *testing.T) {

	p, err := NewProcessorFromFile("tests/target.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	targ := <-p.Targets
	if targ.GetName() == "libxstring" {
		t.Fail()
	}

}

func TestTargetFromMacro(t *testing.T) {
	copts := []string{
		"-std=c11",
		"-fasm",
		"-c",
		"-ffreestanding",
		"-fno-builtin",
		"-fno-omit-frame-pointer",
		"-fplan9-extensions",
		"-fvar-tracking",
		"-fvar-tracking-assignments",
		"-g",
		"-gdwarf-2",
		"-ggdb",
		"-mcmodel=small",
		"-mno-red-zone",
		"-O0",
		"-static",
		"-Wall",
		"-Wno-missing-braces",
		"-Wno-parentheses",
		"-Wno-unknown-pragmas",
	}
	includes := []string{
		"//sys/include",
		"//amd64/include",
	}
	p, err := NewProcessorFromFile("tests/targetFromMacro.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	targ := <-p.Targets
	cBin := targ.(*cc.CLib)
	if cBin.Name == "libxstring" {
		t.Fail()
	}
	for i, v := range copts {
		if cBin.CompilerOptions[i] != v {
			t.Fail()
		}
	}
	for i, v := range includes {
		if cBin.Includes[i] != v {
			t.Fail()
		}
	}

}

func TestTargetFromMacroWithLoad(t *testing.T) {
	copts := []string{
		"-std=c11",
		"-fasm",
		"-c",
		"-ffreestanding",
		"-fno-builtin",
		"-fno-omit-frame-pointer",
		"-fplan9-extensions",
		"-fvar-tracking",
		"-fvar-tracking-assignments",
		"-g",
		"-gdwarf-2",
		"-ggdb",
		"-mcmodel=small",
		"-mno-red-zone",
		"-O0",
		"-static",
		"-Wall",
		"-Wno-missing-braces",
		"-Wno-parentheses",
		"-Wno-unknown-pragmas",
	}
	includes := []string{
		"//sys/include",
		"//amd64/include",
	}
	p, err := NewProcessorFromFile("tests/targetFromMacroWithLoad.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	targ := <-p.Targets
	cBin := targ.(*cc.CLib)
	if cBin.Name == "libxstring" {
		t.Fail()
	}
	for i, v := range copts {
		if cBin.CompilerOptions[i] != v {
			t.Fail()
		}
	}
	for i, v := range includes {
		if cBin.Includes[i] != v {
			t.Fail()
		}
	}

}

func TestTargetFromMacroWithDoubleLoad(t *testing.T) {
	copts := []string{
		"-std=c11",
		"-fasm",
		"-c",
		"-ffreestanding",
		"-fno-builtin",
		"-fno-omit-frame-pointer",
		"-fplan9-extensions",
		"-fvar-tracking",
		"-fvar-tracking-assignments",
		"-g",
		"-gdwarf-2",
		"-ggdb",
		"-mcmodel=small",
		"-mno-red-zone",
		"-O0",
		"-static",
		"-Wall",
		"-Wno-missing-braces",
		"-Wno-parentheses",
		"-Wno-unknown-pragmas",
	}
	includes := []string{
		"//sys/include",
		"//amd64/include",
	}
	p, err := NewProcessorFromFile("tests/targetFromMacroWithDoubleLoadONE.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	targ := <-p.Targets
	cBin := targ.(*cc.CLib)
	if cBin.Name == "libxstring" {
		t.Fail()
	}
	for i, v := range copts {
		if cBin.CompilerOptions[i] != v {
			t.Fail()
		}
	}
	for i, v := range includes {
		if cBin.Includes[i] != v {
			t.Fail()
		}
	}

}

func TestSliceStringVar(t *testing.T) {
	p, err := NewProcessorFromFile("tests/sliceStringVar.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	targ := <-p.Targets
	if targ.GetName() != "test" {
		t.Fail()
	}
}

func TestSliceStringWithStartVar(t *testing.T) {
	p, err := NewProcessorFromFile("tests/sliceStringWithStartVar.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	targ := <-p.Targets
	if targ.GetName() != "test" {
		t.Fail()
	}
}

func TestSliceStringWithEndVar(t *testing.T) {
	p, err := NewProcessorFromFile("tests/sliceStringWithEndVar.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	targ := <-p.Targets
	if targ.GetName() != "lib" {
		t.Fail()
	}
}

func TestArrayIndex(t *testing.T) {
	p, err := NewProcessorFromFile("tests/arrayIndex.BUILD")
	if err != nil {
		t.Fatal(err)
	}
	go p.Run()
	targ := <-p.Targets

	if targ.GetName() != "help" {
		t.Fail()
	}
}

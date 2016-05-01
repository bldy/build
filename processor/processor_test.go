// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package processor // import "sevki.org/build/processor"
import "testing"

//
//func TestLoad(t *testing.T) {
//	p, err := NewProcessorFromFile("tests/load.BUILD")
//	if err != nil {
//		t.Fatal(err)
//	}
//	go p.Run()
//	<-p.Targets
//}

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


func TestMapAssignment(t *testing.T) {
	slc := map[string]string { 
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

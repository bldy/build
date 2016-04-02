// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ast defines build data structures.
package ast // import "sevki.org/build/ast"

import (
	"fmt"
	"log"

	"reflect"
)

var (
	targets map[string]reflect.Type
)

func init() {
	targets = make(map[string]reflect.Type)

}
type File struct {
	Path  string
	Funcs []*Func
	Vars  map[string]interface{}
}

// Node defines what 
type Node struct {
	File           string
	Line, Position int
}

// Variable type points to a variable in, or a loaded document. 
type Variable struct {
	Key string
	Node
}

// Func represents a function in the ast mostly in the form of 
//
// 	glob("", exclude=[], exclude_directories=1)
//
// a function can have named and anonymouse variables at the same time.
type Func struct {
	Name       string
	Params     map[string]interface{}
	AnonParams []interface{}
	Parent     *Func `json:"-"`
	Node
}

// Register function is used to register new types of targets.
func Register(name string, t interface{}) error {
	ty := reflect.TypeOf(t)
	if _, build := reflect.PtrTo(reflect.TypeOf(t)).MethodByName("Build"); !build {
		return fmt.Errorf("%s doesn't implement Build.", reflect.TypeOf(t))
	}
	targets[name] = ty

	return nil
}

// Get returns a reflect.Type for a given name.
func Get(name string) reflect.Type {
	if t, ok := targets[name]; ok {
		return t
	} else {
		log.Fatalf("unregistered target type %s", name)
		return nil
	}
}

// GetFieldByTag returns field by tag
func GetFieldByTag(tn, tag string, p reflect.Type) (*reflect.StructField, error) {
	if p == nil {
		return nil, fmt.Errorf("%s isn't a registered type.", tn)
	}

	for i := 0; i < p.NumField(); i++ {
		f := p.Field(i)
		if f.Tag.Get(tn) == tag {
			return &f, nil
		}
	}
	return nil, fmt.Errorf("%s isn't a field of %s.", tag, tn)
}

// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ast defines build data structures.
package ast // import "sevki.org/build/ast"

import (
	"fmt"
	"io"

	"reflect"
)

var (
	targets map[string]reflect.Type
)

func init() {
	targets = make(map[string]reflect.Type)

}

type BuildInfo struct {
	BuildDir string
	OutDir   string
}

type BuildFile struct {
	Path  string
	Funcs []*Func
	Vars  map[string]interface{}
}
type Varuable struct {
	Value string
}
type Func struct {
	Name       string
	Params     map[string]interface{}
	AnonParams []interface{}
	Parent     *Func `json:"-"`
}

// Target is for implementing different build targets.
type Target interface {
	Build() error
	Install() error
	GetName() string
	GetDependencies() []string
	GetSource() string
	Reader() io.Reader
	Hash() []byte
}

type Path string

func Register(name string, t interface{}) error {
	ty := reflect.TypeOf(t)
	if _, build := reflect.PtrTo(reflect.TypeOf(t)).MethodByName("Build"); !build {
		return fmt.Errorf("%s doesn't implement Build.", reflect.TypeOf(t))
	}
	targets[name] = ty

	return nil
}
func Get(name string) reflect.Type {
	if t, ok := targets[name]; ok {
		return t
	} else {
		return nil
	}
}
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

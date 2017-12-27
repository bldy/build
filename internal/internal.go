// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package internal is used for registering types in build, it had no clear place
// in other packages to go which is why it gets it's own package
package internal

import (
	"fmt"

	"reflect"
)

var (
	rules map[string]reflect.Type
)

func init() {
	rules = make(map[string]reflect.Type)
}

// Register function is used to register new types of targets.
func Register(name string, t interface{}) error {
	ty := reflect.TypeOf(t)
	if _, build := reflect.PtrTo(reflect.TypeOf(t)).MethodByName("Build"); !build {
		return fmt.Errorf("%s doesn't implement Build", reflect.TypeOf(t))
	}
	rules[name] = ty

	return nil
}

// Get returns a reflect.Type for a given name.
func Get(name string) reflect.Type {
	if t, ok := rules[name]; ok {
		return t
	}
	return nil
}

// GasdetFieldByTag returns field by tag
func GasdetFieldByTag(tn, tag string, p reflect.Type) (*reflect.StructField, error) {
	if p == nil {
		return nil, fmt.Errorf("%s isn't a registered type", tn)
	}

	for i := 0; i < p.NumField(); i++ {
		f := p.Field(i)
		if f.Tag.Get(tn) == tag {
			return &f, nil
		}
	}
	return nil, fmt.Errorf("%s isn't a field of %s", tag, tn)
}

// Rules returns registered RulesTypes
func Rules() []string {
	s := []string{}
	for k := range rules {
		s = append(s, k)
	}
	return s
}

package skylark

import (
	"github.com/google/skylark"
	"github.com/pkg/errors"
)

// newAttributor returns a new bazel build context.
func attributors() bldyDict {
	attributors := make(bldyDict)
	for _, actionName := range []string{
		"bool",
		"int",
		"int_list",
		"label",
		"label_keyed_string_dict",
		"label_list",
		"license",
		"output",
		"output_list",
		"string",
		"string_dict",
		"string_list",
		"string_list_dict",
	} {
		attributors[actionName] = attributer{actionName}
	}
	return attributors
}

type attributer struct {
	attrType string
}

func (a attributer) Name() string          { return a.attrType }
func (a attributer) Hash() (uint32, error) { return hashString(a.attrType), nil }
func (a attributer) Freeze()               {}
func (a attributer) String() string        { return a.attrType }
func (a attributer) Type() string          { return "attributer" }
func (a attributer) Truth() skylark.Bool   { return true }
func (a attributer) Call(thread *skylark.Thread, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var i Attribute
	x := attr{attrType: a.attrType}
	switch a.attrType {
	case "bool":
		i = &x
	case "int":
		i = &intAttr{attr: x}
	case "int_list":
		i = &intListAttr{attr: x}
	case "label":
		i = &labelAttr{attr: x}
	case "label_keyed_string_dict":
		i = &labelKeyedStringDictAttr{attr: x}
	case "label_list":
		i = &labelListAttr{attr: x}
	case "license",
		"output",
		"output_list",
		"string",
		"string_dict",
		"string_list",
		"string_list_dict":
		panic("not implemented")
	}
	if err := unpackStruct(i, kwargs); err != nil {
		return nil, errors.Wrap(err, "attiributor.call")
	}
	return i, nil
}

// Attribute is representation of a definition of an attribute.
// Use the attr module to create an Attribute.
// They are only for use with a rule or an aspect.
// https://docs.bazel.build/versions/master/skylark/lib/Attribute.html
type Attribute skylark.Value

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#modules.attr
type attr struct {
	attrType string

	// Common to all Attrs
	Default   skylark.Value
	Doc       string
	Mandatory bool
}

func (a *attr) Name() string          { return a.attrType }
func (a *attr) Hash() (uint32, error) { return hashString(a.attrType), nil }
func (a *attr) Freeze()               {}
func (a *attr) String() string        { return a.attrType }
func (a *attr) Type() string          { return "attr." + a.attrType }
func (a *attr) Truth() skylark.Bool   { return true }

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#bool
type boolAttr struct {
	attr
}

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#int
type intAttr struct {
	attr

	Values []int
}

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#int_list
type intListAttr struct {
	attr

	NonEmpty   bool
	AllowEmpty bool
}

type configuration string

const (
	Data   configuration = "data"
	Host   configuration = "host"
	Target configuration = "target"
)

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#label
type labelAttr struct {
	attr

	Executable           bool
	AllowFiles           bool
	AllowSingleFile      bool
	AllowdExtensionsList []string
	Providers            [][]string

	SingleFile bool

	Cfg configuration
}

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#label_keyed_string_dict
type labelKeyedStringDictAttr struct {
	attr

	Executable           bool
	AllowFiles           bool
	AllowdExtensionsList []string
	Providers            [][]string

	SingleFile bool

	Cfg configuration
}

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#label_list
type labelListAttr struct {
	attr

	Executable           bool
	AllowFiles           bool
	AllowdExtensionsList []string
	Providers            [][]string

	SingleFile bool

	Cfg configuration
}

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#license
type licenseAttr struct{ attr }

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#output
type outputAttr struct{ attr }

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#output_list
type outputListAttr struct {
	attr

	NonEmpty   bool
	AllowEmpty bool
}

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#string
type stringAttr struct {
	attr

	Values []string
}

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#string_dict
type stringDictAttr struct {
	attr

	NonEmpty   bool
	AllowEmpty bool
}

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#string_list
type stringListAttr struct {
	attr

	NonEmpty   bool
	AllowEmpty bool
}

// https://docs.bazel.build/versions/master/skylark/lib/attr.html#string_list_dict
type stringListDict struct {
	attr

	NonEmpty   bool
	AllowEmpty bool
}

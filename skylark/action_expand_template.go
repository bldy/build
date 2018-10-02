package skylark

import (
	"fmt"
	"io/ioutil"
	"text/template"

	"bldy.build/build/skylark/skylarkutils"

	"bldy.build/build/executor"
	"bldy.build/build/file"
	"github.com/google/skylark"
	"github.com/pkg/errors"
)

// Expand Template represents a ctx.actions.run functions in bazel land.
// https://docs.bazel.build/versions/master/skylark/lib/actions.html#run
type expand_template struct {
	Template      *file.File    // The template file, which is a UTF-8 encoded text file.
	Output        *file.File    // The output file, which is a UTF-8 encoded text file.
	Substitutions *skylark.Dict // Substitutions to make when expanding the template.
	IsExecutable  bool          // Whether the output file should be executable.
}

var (
	funcs = template.FuncMap{
		"as_string": func(v interface{}) string {
			val, ok := v.(skylark.Value)
			if !ok {
				return fmt.Sprintf("!(INVALID %T %s)", v, v)
			}
			s, ok := skylark.AsString(val)
			if !ok {
				return fmt.Sprintf("!(INVALID %T %s)", v, v)
			}
			return s
		},
		"as_slice": func(v interface{}) skylark.Tuple {
			val, ok := v.(*skylark.List)
			if !ok {
				return []skylark.Value{skylark.String(fmt.Sprintf("!(INVALID %T %s)", v, v))}
			}
			s, err := skylarkutils.ListToTuple(val)
			if err != nil {
				return []skylark.Value{skylark.String(fmt.Sprintf("!(INVALID %T %s)", v, v))}
			}
			return s
		},
	}
)

func (et *expand_template) Do(e *executor.Executor) error {
	bytz, err := ioutil.ReadFile(et.Template.Path())
	if err != nil {
		return errors.Wrap(err, "expand_template: read file")
	}
	tmpl, err := template.New("root").Funcs(funcs).Parse(string(bytz))
	if err != nil {
		return errors.Wrap(err, "expand_template")
	}

	m, err := skylarkutils.DictToStringDict(et.Substitutions)
	if err != nil {
		return errors.Wrap(err, "expand_template: substitutions to go")
	}
	f, err := e.Create(et.Output.Path())
	if err != nil {
		return errors.Wrap(err, "expand_template: creating out file")
	}
	return tmpl.Funcs(funcs).Execute(f, m)
}

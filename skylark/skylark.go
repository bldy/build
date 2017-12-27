package skylark

import (
	"fmt"
	"log"
	"os"

	"bldy.build/build/label"

	"bldy.build/build"
	"github.com/google/skylark"
	"github.com/pkg/errors"
)

const (
	skylarkKeyImpl   = "implementation"
	skylarkKeyAttrs  = "attrs"
	skylarkKeyDeps   = "deps"
	skylarkKeyName   = "name"
	threadKeyTargets = "__targets"
	threadKeyWD      = "__wd"
	threadKeyContext = "__ctx"
	threadKeyPackage = "__package"
)

var (
	l = log.New(os.Stdout, "skylark: ", log.Lshortfile)
)

func init() {
	skylark.Universe["attr"] = attributer{}
}

type skylarkVM struct {
	wd    string
	pkg   string
	rules []build.Rule
}

// New returns a new skylarkVM
func New(wd string) (build.VM, error) {
	return &skylarkVM{
		wd: wd,
	}, nil
}

func print(thread *skylark.Thread, msg string) {
	l.Println(msg)
}

func (s *skylarkVM) GetTarget(l *label.Label) (build.Rule, error) {
	bytz, err := label.LoadLabel(l)
	if err != nil {
		errors.Wrap(err, "get target:")
	}

	t := &skylark.Thread{}
	t.Load = s.load
	t.Print = print

	t.SetLocal(threadKeyPackage, l.Package)

	globals := skylark.StringDict{
		"rule": skylark.NewBuiltin("rule", s.makeRule),
	}
	err = skylark.ExecFile(t, l.String(), bytz, globals)
	if err != nil {
		return nil, errors.Wrap(err, "skylark: eval")
	}
	for _, r := range s.rules {
		if r.GetName() == l.Name {
			return r, nil
		}
	}
	return nil, fmt.Errorf("couldn't find the target %s", l.String())
}

func (s *skylarkVM) load(thread *skylark.Thread, module string) (skylark.StringDict, error) {
	bytz, err := label.Load(module)
	if err != nil {
		log.Println(err)
		return nil, errors.Wrap(err, "skylark: eval")
	}
	globals := skylark.StringDict{
		"rule": skylark.NewBuiltin("rule", s.makeRule),
	}
	err = skylark.ExecFile(thread, module, bytz, globals)
	if err != nil {
		return globals, err
	}
	return globals, nil
}

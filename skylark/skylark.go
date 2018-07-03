package skylark

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"bldy.build/build/internal"
	"bldy.build/build/label"
	"bldy.build/build/workspace"

	"bldy.build/build"
	"github.com/google/skylark"
	"github.com/google/skylark/skylarkstruct"
	"github.com/pkg/errors"
)

const (
	skylarkKeyImpl    = "implementation"
	skylarkKeyAttrs   = "attrs"
	skylarkKeyDeps    = "deps"
	skylarkKeyOutputs = "outputs"
	skylarkKeyName    = "name"
	threadKeyTargets  = "__targets"
	threadKeyWD       = "__wd"
	threadKeyContext  = "__ctx"
	threadKeyPackage  = "__package"
)

var (
	l             = log.New(os.Stdout, "skylark: ", log.Lshortfile)
	errDoesntHash = errors.New("doesn't implement hash")
)

func init() {
	skylark.Universe["attr"] = attributors()
	skylark.Universe["env"] = skylark.NewBuiltin("env", _env)
}

type skylarkVM struct {
	pkg     string
	rules   map[string]build.Rule
	ws      workspace.Workspace
	globals skylark.StringDict
}

// New returns a new skylarkVM
func New(ws workspace.Workspace) (build.VM, error) {
	s := &skylarkVM{
		ws:    ws,
		rules: make(map[string]build.Rule),
	}

	natives := make(nativeMap)
	for _, nativeRule := range internal.Rules() {
		natives[nativeRule] = skylark.NewBuiltin(nativeRule, s.makeNativeRule)
	}
	globals := skylark.StringDict{
		"rule":   skylark.NewBuiltin("rule", s.makeRule),
		"native": natives,
		"struct": skylark.NewBuiltin("struct", skylarkstruct.Make),
	}
	s.globals = globals
	return s, nil
}

func print(thread *skylark.Thread, msg string) {
	l.Println("something something ", msg)
}

func (s *skylarkVM) GetPackageDir(l *label.Label) string {
	return s.ws.PackageDir(l)

}
func (s *skylarkVM) GetTarget(l *label.Label) (build.Rule, error) {
	if r, ok := s.rules[l.String()]; ok {
		return r, nil
	}

	bytz, err := s.ws.LoadBuildfile(l)
	if err != nil {
		errors.Wrap(err, "skylark.get_target:")
	}

	t := &skylark.Thread{}
	t.Load = s.load
	t.Print = print

	t.SetLocal(threadKeyPackage, *l.Package)

	if _, err = skylark.ExecFile(t, s.ws.Buildfile(l), bytz, s.globals); err != nil {
		return nil, errors.Wrap(err, "skylark.exec")
	}
	if r, ok := s.rules[l.String()]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("skylark: couldn't find the target %s in %s", l.String(), s.ws.Buildfile(l))
}

func (s *skylarkVM) load(thread *skylark.Thread, module string) (skylark.StringDict, error) {

	lbl, err := label.Parse(module)
	if lbl.Package == nil {
		lbl.Package = label.Package(thread.Local(threadKeyPackage).(string))
	}

	bytz, err := s.ws.LoadBuildfile(lbl)
	if err != nil {
		buf := bytes.NewBuffer(nil)
		thread.Caller().WriteBacktrace(buf)
		return nil, fmt.Errorf("skylark.load: %s\n%s", err.Error(), buf.String())
	}

	return skylark.ExecFile(thread, s.ws.Buildfile(lbl), bytz, s.globals)
}

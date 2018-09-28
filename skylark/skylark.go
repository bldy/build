package skylark

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"bldy.build/build/internal"
	"bldy.build/build/label"
	"bldy.build/build/workspace"
	"sevki.org/x/debug"

	"bldy.build/build"
	"github.com/google/skylark"
	"github.com/google/skylark/skylarkstruct"
	"github.com/pkg/errors"
)

const (
	skylarkKeyImpl           = "implementation"
	skylarkKeyAttrs          = "attrs"
	skylarkKeyDeps           = "deps"
	skylarkKeyOutputs        = "outputs"
	skylarkKeyName           = "name"
	skylarkKeyCompatibleWith = "compatible_with"
	skylarkKeyToolChains     = "toolchains"
	skylarkKeyHost           = "host"
	skylarkKeyRestrictedTo   = "restricted_to"
	skylarkKeyTags           = "tags"

	threadKeyTargets = "__targets"
	threadKeyWD      = "__wd"
	threadKeyContext = "__ctx"
	threadKeyPackage = "__package"
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

	natives := skylark.StringDict{}
	for _, nativeRule := range internal.Rules() {
		natives[nativeRule] = skylark.NewBuiltin(nativeRule, s.makeNativeRule)
	}
	globals := skylark.StringDict{
		"rule":   skylark.NewBuiltin("rule", s.makeRule),
		"glob":   skylark.NewBuiltin("glob", s.glob),
		"native": skylarkstruct.FromStringDict(skylarkstruct.Default, natives),
		"struct": skylark.NewBuiltin("struct", skylarkstruct.Make),
	}
	s.globals = globals
	return s, nil
}

func print(thread *skylark.Thread, msg string) {
	l.Println("something something ", msg)
}

func (s *skylarkVM) GetPackageDir(l label.Label) string {
	return s.ws.PackageDir(l)

}
func (s *skylarkVM) GetTarget(l label.Label) (build.Rule, error) {

	if r, ok := s.rules[l.String()]; ok {
		return r, nil
	}

	bytz, err := s.ws.LoadBuildfile(l)
	if err != nil {
		return nil, errors.Wrap(err, "skylark.get_target:")
	}
	if err := l.Valid(); err != nil {
		return nil, errors.Wrap(err, "skylark.get_target:")
	}
	if l.Package() == "" {
		return nil, errors.New("skylark vm can't figure out labels without packages, for the root package please use '.'.")
	}

	t := &skylark.Thread{}
	t.Load = s.load
	t.Print = print
	initPkgStack(t)
	pushPkg(t, l.Package())

	if _, err = skylark.ExecFile(t, s.ws.Buildfile(l), bytz, s.globals); err != nil {
		return nil, errors.Wrap(err, "skylark: gettarget: exec")
	}
	if r, ok := s.rules[l.String()]; ok {
		r.(*Rule).ws = s.ws // TODO: fix this

		return r, nil
	}

	return nil, fmt.Errorf("skylark: couldn't find the target %q in %s", l, s.ws.Buildfile(l))
}

func (s *skylarkVM) load(thread *skylark.Thread, module string) (skylark.StringDict, error) {
	pkg := getPkg(thread)
	l, err := label.Parse(module)

	if err != nil {
		return nil, errors.Wrap(err, "skylark.load")
	}
	if !l.IsAbs() {
		l = label.New(pkg, module)
	}

	file := ""
	if path.Ext(l.Name()) != "" {
		file = s.ws.File(l)
	} else {
		file = s.ws.Buildfile(l)
	}
	bytz, err := ioutil.ReadFile(file)
	if err != nil {
		buf := bytes.NewBuffer(nil)
		thread.Caller().WriteBacktrace(buf)
		return nil, fmt.Errorf("skylark.load: %s\n%s", err.Error(), buf.String())
	}

	pushPkg(thread, l.Package())
	dict, err := skylark.ExecFile(thread, s.ws.Buildfile(l), bytz, s.globals)
	if err != nil {
		buf := bytes.NewBuffer(nil)
		debug.Indent(buf, 1)
		thread.Caller().WriteBacktrace(buf)
		return nil, fmt.Errorf("skylark: load: exec: %s\n%s", err.Error(), buf.String())
	}
	popPkg(thread)
	return dict, err
}

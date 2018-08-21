package skylark

import (
	"encoding/hex"
	"strings"

	"bldy.build/build/deb"
	"bldy.build/build/label"
	"github.com/google/skylark"
)

func (s *skylarkVM) makeDebRule(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	pkg := getPkg(thread)

	var name skylark.String
	var arch, release, repo, component, debpkg, version skylark.String
	var checksum skylark.String
	var deps *skylark.List
	var outs *skylark.List

	_ = deps

	err := skylark.UnpackArgs("debian rule", args, kwargs,
		skylarkKeyName, &name,
		skylarkKeyOutputs+"?", &outs,
		"pkg?", &debpkg,
		"checksum?", &checksum,
		"repo", &repo,
		"release", &release,
		"arch", &arch,
		"component", &component,
		"version", &version,
	)
	if err != nil {
		return skylark.None, err
	}
	if string(debpkg) == "" {
		debpkg = name
	}
	lbl := label.New(pkg, string(name))
	var hash []byte

	outputs := []string{}

	if outs != nil {
		if o, err := skylarkToGo(outs); err == nil {
			if oo, ok := o.([]string); ok {
				outputs = oo
			}
		} else {
			panic(err)
		}
	}

	if i := strings.IndexByte(string(checksum), ':'); i > 3 {
		if decoded, err := hex.DecodeString(string(checksum)[i+1:]); err != nil {
			return skylark.None, err
		} else {
			hash = decoded
		}
	}
	s.rules[lbl.String()] = deb.NewDebianRule(
		string(name),
		string(repo),
		string(release),
		string(component),
		string(arch),
		string(version),
		string(debpkg),
		hash,
		outputs,
	)
	return skylark.None, nil
}

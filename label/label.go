// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package label // import "bldy.build/build/label"
import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bldy.build/build/project"
	"github.com/pkg/errors"
)

const buildfile = "BUILD"

// Label represents a perforce label
// we plan on adding more providers
type Label struct {
	Package string
	Name    string
}

// LoadLabel takes a label and returns the contents of it.
func LoadLabel(lbl *Label) ([]byte, error) {
	buildpath := lbl.BuildFile()
	_, err := os.Stat(buildpath)
	if err != nil {
		return nil, fmt.Errorf("label.load: file %q doesn't exist", buildpath)
	}
	bytz, err := ioutil.ReadFile(buildpath)
	if err != nil {
		return nil, errors.Wrap(err, "skylark: readall")
	}
	return bytz, nil
}

// Load takes a string, which can be a label or a filepath and returns
// the contents of it.
func Load(s string) ([]byte, error) {
	ext := path.Ext(s)
	if ext != "" {
		bytz, err := ioutil.ReadFile(s)
		if err != nil {
			return nil, errors.Wrap(err, "label.load: readall")
		}
		return bytz, nil
	}
	lbl, err := Parse(s)
	if err != nil {
		return nil, errors.Wrap(err, "label.load")
	}
	return LoadLabel(lbl)
}

func split(s string, c string, cutc bool) (string, string) {
	i := strings.Index(s, c)
	if i < 0 {
		return s, ""
	}
	if cutc {
		return s[:i], s[i+len(c):]
	}
	return s[:i], s[i:]
}

func (lbl Label) String() string {
	return fmt.Sprintf("//%s:%s", lbl.Package, lbl.Name)
}

// BuildDir Returns the BuildDir for a given label
// it takes two args, workdir and project
func (lbl Label) BuildFile() string {
	wd, _ := os.Getwd()
	project := project.GetGitDir(wd)
	if lbl.Package == "" {
		return wd
	}
	ext := path.Ext(lbl.Name)
	if ext != "" {
		return filepath.Join(project, lbl.Package, lbl.Name)
	}
	return filepath.Join(project, lbl.Package, "BUILD")
}

// Parse takes a string and returns a bazel label
func Parse(s string) (*Label, error) {
	lbl := new(Label)
	switch {
	case strings.HasPrefix(s, "//"):
		s = s[2:]
		lbl.Package, lbl.Name = split(s, ":", true)
		if lbl.Package == "" {
			lbl.Package = "." // this is the root of the project
		}
	case strings.HasPrefix(s, ":"):
		s = s[1:]
		fallthrough
	default:
		lbl.Name = s
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		lbl.Package, err = filepath.Rel(project.Root(), wd)
		if err != nil {
			log.Fatal(err)
		}
	}
	if lbl.Name == "" {
		lbl.Name = path.Base(lbl.Package)
	}

	return lbl, nil
}

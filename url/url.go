// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package url // import "bldy.build/build/url"
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

// URL represents a perforce URL
// we plan on adding more providers
type URL struct {
	Package string
	Target  string
}

// LoadURL takes a URL, which will be a perforce url returns the contents of it.
func LoadURL(u *URL) ([]byte, error) {
	buildpath := path.Join(u.BuildDir(), buildfile)
	_, err := os.Stat(buildpath)
	if err != nil {
		return nil, fmt.Errorf("url: load: file %q doesn't exist", buildpath)
	}
	bytz, err := ioutil.ReadFile(buildpath)
	if err != nil {
		return nil, errors.Wrap(err, "skylark: readall")
	}
	return bytz, nil
}

// Load takes a string, which can be a perforce url or a filepath and returns
// the contents of it.
func Load(s string) ([]byte, error) {
	ext := path.Ext(s)
	if ext != "" {
		bytz, err := ioutil.ReadFile(s)
		if err != nil {
			return nil, errors.Wrap(err, "load: readall")
		}
		return bytz, nil
	}
	u, err := Parse(s)
	if err != nil {
		return nil, errors.Wrap(err, "load")
	}
	return LoadURL(u)
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

func (u URL) String() string {
	return fmt.Sprintf("//%s:%s", u.Package, u.Target)
}

// BuildDir Returns the BuildDir for a given url
// it takes two args, workdir and project
func (u URL) BuildDir() string {
	wd, _ := os.Getwd()
	project := project.GetGitDir(wd)
	if u.Package == "" {
		return wd
	}
	return filepath.Join(project, u.Package)
}

func Parse(s string) (*URL, error) {
	u := new(URL)
	switch {
	case strings.HasPrefix(s, "//"):
		s = s[2:]
		u.Package, u.Target = split(s, ":", true)
		if u.Package == "" {
			u.Package = "." // this is the root of the project
		}
	case strings.HasPrefix(s, ":"):
		s = s[1:]
		fallthrough
	default:
		u.Target = s
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		u.Package, err = filepath.Rel(project.Root(), wd)
		if err != nil {
			log.Fatal(err)
		}
	}
	if u.Target == "" {
		u.Target = path.Base(u.Package)
	}

	return u, nil
}

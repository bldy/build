// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package url // import "bldy.build/build/url"
import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bldy.build/build/util"
)

type URL struct {
	Package string
	Target  string
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

func (u URL) BuildDir(wd, p string) string {
	if u.Package == "" {
		return wd
	} else {
		return filepath.Join(p, u.Package)
	}
}
func Parse(s string) (u URL) {

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
		u.Package, err = filepath.Rel(util.GetProjectPath(), wd)
		if err != nil {
			log.Fatal(err)
		}
	}
	if u.Target == "" {
		u.Target = path.Base(u.Package)
	}

	return
}

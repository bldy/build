// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harvey

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/nu7hatch/gouuid"

	"bldy.build/build"
)

type OldBuild struct {
	Name         string   `old_build:"name"`
	Dependencies []string `old_build:"deps"`
	Package      string   `old_build:"package"`
}

func (ob *OldBuild) GetName() string {
	return ob.Name
}

func (ob *OldBuild) GetDependencies() []string {
	return ob.Dependencies
}

func (s *OldBuild) Hash() []byte {
	h := sha1.New()
	u, _ := uuid.NewV4()
	io.WriteString(h, u.String())
	return []byte{}
}

func oldbuild() string {
	return path.Join(os.Getenv("HARVEY"), "util/build")
}
func (s *OldBuild) Build(e *build.Executor) error {
	params := []string{s.Package}

	if err := e.Exec(oldbuild(), nil, params); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}
func (s *OldBuild) Installs() map[string]string {
	return nil
}

// +build !test

package testws

import (
	"path/filepath"

	"bldy.build/build/label"
)

type TestWS struct {
	WD string
}

func (t *TestWS) AbsPath() string {
	panic("not implemented")
}

func (t *TestWS) Buildfile(label.Label) string {
	panic("not implemented")
}

func (t *TestWS) File(lbl label.Label) string {
	pkg, name, err := lbl.Split()
	if err != nil {
		panic(err)
	}
	return filepath.Join(t.WD, pkg, name)
}

func (t *TestWS) PackageDir(lbl label.Label) string {

	return filepath.Join(t.WD, lbl.Package())
}

func (t *TestWS) LoadBuildfile(label.Label) ([]byte, error) {
	panic("not implemented")
}

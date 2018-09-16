package workspace

import (
	"io/ioutil"
	"path/filepath"

	"bldy.build/build/label"
)

type localWorkspace struct {
	wd string
}

func (lw *localWorkspace) AbsPath() string {
	return lw.wd
}
func (lw *localWorkspace) PackageDir(lbl label.Label) string {
	if err := lbl.Valid(); err != nil {
		panic(err)
	}
	return filepath.Join(lw.wd, lbl.Package())
}

func (lw *localWorkspace) File(lbl label.Label) string {
	if err := lbl.Valid(); err != nil {
		panic(err)
	}
	return filepath.Join(lw.wd, lbl.Package(), lbl.Name())
}

func (lw *localWorkspace) Buildfile(lbl label.Label) string {
	return filepath.Join(lw.PackageDir(lbl), BUILDFILE)
}

func (lw *localWorkspace) LoadBuildfile(lbl label.Label) ([]byte, error) {
	file := lw.Buildfile(lbl)
	return ioutil.ReadFile(file)
}

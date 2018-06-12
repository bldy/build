package workspace

import (
	"io/ioutil"
	"path"
	"path/filepath"

	"bldy.build/build/label"
)

type localWorkspace struct {
	wd string
}

func (lw *localWorkspace) AbsPath() string {
	return lw.wd
}
func (lw *localWorkspace) PackageDir(lbl *label.Label) string {
	var pkg string
	if lbl.Package == nil {
		pkg = lw.wd
	} else {
		pkg = *lbl.Package
	}

	return filepath.Join(lw.wd, pkg)

}
func (lw *localWorkspace) Buildfile(lbl *label.Label) string {
	var pkg string
	if lbl.Package == nil {
		pkg = lw.wd
	} else {
		pkg = *lbl.Package
	}
	ext := path.Ext(lbl.Name)
	if ext != "" {
		return filepath.Join(lw.wd, pkg, lbl.Name)
	}
	return filepath.Join(lw.wd, pkg, BUILDFILE)
}

func (lw *localWorkspace) LoadBuildfile(lbl *label.Label) ([]byte, error) {
	file := lw.Buildfile(lbl)
	return ioutil.ReadFile(file)
}

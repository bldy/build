package workspace

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"bldy.build/build/label"
	"github.com/pkg/errors"
)

const (
	// BUILDFILE is the name of the file that keeps targets
	BUILDFILE = "BUILD"
)

var (
	// ErrNotAWorkspace is returned when the given URL is not a workspace
	ErrNotAWorkspace = errors.New("workspace: given url is not a workspace")
	// ErrNotAbsolute is returned when a url is not absolute
	ErrNotAbsolute = errors.New("workspace: local urls should be absolute paths")
)

// Workspace is a Bazel workspace
// https://docs.bazel.build/versions/master/be/workspace.html
type Workspace interface {
	AbsPath() string
	Buildfile(*label.Label) string
	LoadBuildfile(*label.Label) ([]byte, error)
}

// New given a URL returns a workspace
func New(a string) (Workspace, error) {
	u, err := url.Parse(a)
	if err != nil {
		return nil, errors.Wrap(err, "workspace new")
	}
	if u.Scheme == "" {
		if !path.IsAbs(a) {
			return nil, ErrNotAbsolute
		}
		wd, err := FindWorkspace(a, os.Lstat)
		if err != nil {
			return nil, err
		}
		return &localWorkspace{
			wd: wd,
		}, nil
	}

	return nil, ErrNotAWorkspace
}

// Stat checks if a file exists or not in a workspace
type Stat func(string) (os.FileInfo, error)

// FindWorkspace looks for recursively for WORKSPACE file in each
// parent directory. If it fails to find anything it will return the fist
// directory with .git
//
// A workspace is a directory on your filesystem that contains the source files
// for the software you want to build, as well as symbolic links to directories
// that contain the build outputs. Each workspace directory has a text file
// named WORKSPACE which may be empty, or may contain references to external
//dependencies required to build the outputs.
// https://docs.bazel.build/versions/master/build-ref.html#workspace
func FindWorkspace(p string, stat Stat) (string, error) {

	dirs := strings.Split(p, "/")
	for i := len(dirs) - 1; i > 0; i-- {
		frags := append([]string{"/"}, dirs[0:i+1]...)
		path := path.Join(frags...)
		try := fmt.Sprintf("%s/WORKSPACE", path)
		if _, err := stat(try); os.IsNotExist(err) {
			continue
		}
		return path, nil
	}
	return "", ErrNotAWorkspace
}

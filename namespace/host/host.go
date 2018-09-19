package host

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"bldy.build/build/namespace"
	"github.com/pkg/errors"
	"sevki.org/x/debug"
)

type Namespace struct {
	dir string
	env []string
}

func New(name string) (namespace.Namespace, error) {
	if err := os.MkdirAll(name, 0755); err != nil {
		return nil, errors.Wrap(err, "new host namespace")
	}
	return Namespace{dir: name}, nil
}

func (n Namespace) environ() []string {
	env := []string{}
	for key, paths := range map[string][]string{
		"PATH":           []string{"/usr/local/sbin", "/usr/local/bin", "/usr/sbin", "/usr/bin", "/sbin", "/bin", "usr/lib/llvm-3.8/bin/"},
		"C_INCLUDE_PATH": []string{"/usr/local/include", "/usr/include", "/include"},
		"LIBRARY_PATH":   []string{"/usr/local/lib", "/usr/lib", "/lib", "usr/lib/x86_64-linux-gnu"},
	} {
		namespaced := []string{os.Getenv(key)}
		for _, p := range paths {
			namespaced = append(namespaced, path.Join(n.dir, p))
		}
		env = append(env, fmt.Sprintf("%s=%s", key, strings.Join(namespaced, ":")))
	}
	return append(env, n.env...)
}

// Namespace is where builds are run, interface is the same as the plan9 namespaces
func (n Namespace) Bind(new, old string, flags int) {
	if path, err := filepath.EvalSymlinks(old); err == nil && old != path {
		old = path
	}
	if err := os.Symlink(old, new); err != nil {
		debug.Println(err)
	}
}
func (ns Namespace) Mount(new, old string, flags int) {}

func (n Namespace) Cmd(ctx context.Context, cmd string, args ...string) namespace.Cmd {
	x := exec.CommandContext(ctx, cmd, args...)
	x.Env = n.environ()
	x.Dir = n.dir
	return x
}

func (n Namespace) Mkdir(name string) error {
	return os.MkdirAll(filepath.Join(n.dir, name), os.ModeDir|os.ModePerm)
}

func (n Namespace) Open(name string) (*os.File, error) {
	if filepath.IsAbs(name) {
		return os.Open(name)
	}
	return os.Open(filepath.Join(n.dir, name))
}

func (n Namespace) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(filepath.Join(n.dir, name), flag, perm)
}

func (n Namespace) Create(name string) (*os.File, error) {
	return os.Create(filepath.Join(n.dir, name))
}

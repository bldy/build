package harvey

import (
	"crypto/sha1"
	"fmt"
	"io"
	"path/filepath"

	"sevki.org/build"
)

type Strip struct {
	Name         string   `strip:"name"`
	Dependencies []string `strip:"deps"`
	Bin          string   `strip:"bin"`
}

func (s *Strip) GetName() string {
	return s.Name
}

func (s *Strip) GetDependencies() []string {
	return s.Dependencies
}

func (s *Strip) Hash() []byte {
	h := sha1.New()

	io.WriteString(h, s.Name)
	io.WriteString(h, s.Bin)
	return []byte{}
}

// Had to be done
func Stripper() string {
	if tpfx := build.Getenv("TOOLPREFIX"); tpfx == "" {
		return "strip"
	} else {
		return fmt.Sprintf("%s%s", tpfx, "strip")
	}
}
func (s *Strip) Build(c *build.Context) error {
	params := []string{"-o"}
	params = append(params, s.Name)
	params = append(params, filepath.Join("bin", s.Bin))
	if err := c.Exec(Stripper(), nil, params); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}
func (s *Strip) Installs() map[string]string {
	installs := make(map[string]string)
	fname := fmt.Sprintf("%s.c", s.Name)
	installs[filepath.Join("bin", fname)] = fname
	return installs
}

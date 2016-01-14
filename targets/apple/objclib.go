package apple

import (
	"crypto/sha1"
	"io"

	"sevki.org/build"
	"sevki.org/build/util"
)

type ObjCLib struct {
	Name         string   `objc_library:"name" `
	Dependencies []string `objc_library:"deps"`
	Sources      []string `objc_library:"srcs" build:"path"`
	Headers      []string `objc_library:"hdrs" build:"path"`
	XIBs         []string `objc_library:"xibs" build:"path"`
}

func (ol *ObjCLib) GetName() string {
	return ol.Name
}
func (ol *ObjCLib) GetDependencies() []string {
	return append(ol.Dependencies)
}

func (ol *ObjCLib) Hash() []byte {
	h := sha1.New()
	io.WriteString(h, ol.Name)
	util.HashFiles(h, ol.Sources)
	util.HashFiles(h, ol.XIBs)
	util.HashFiles(h, ol.Headers)
	return h.Sum(nil)
}

func (ol *ObjCLib) Build(c *build.Context) error {
	return nil
}
func (ol *ObjCLib) Installs() map[string]string {
	return nil
}

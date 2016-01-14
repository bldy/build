package apple

import (
	"crypto/sha1"
	"fmt"
	"io"

	"sevki.org/build"
	"sevki.org/build/util"
)

type IOSApplication struct {
	Name         string   `ios_application:"name" `
	Dependencies []string `ios_application:"deps"`
	Binary       string   `ios_application:"binary"`
	InfoPlist    string   `ios_application:"infoplist" build:"path"`
}

func (ia *IOSApplication) GetName() string {
	return ia.Name
}
func (ia *IOSApplication) GetDependencies() []string {
	return append([]string{fmt.Sprintf(":%s", ia.Binary)}, ia.Dependencies...)
}

func (ia *IOSApplication) Hash() []byte {
	h := sha1.New()
	io.WriteString(h, ia.Name)
	util.HashFiles(h, []string{ia.InfoPlist})
	return h.Sum(nil)
}

func (ia *IOSApplication) Build(c *build.Context) error {
	return nil
}
func (ia *IOSApplication) Installs() map[string]string {
	return nil
}

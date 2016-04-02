package harvey

import (
	"crypto/sha1"

	"io"

	"sevki.org/build"
)

type Kernel struct {
	Name         string            `kernel:"name"`
	Dependencies []string          `kernel:"deps"`
	RamFiles     []string          `kernel:"ramfiles" build:"path"`
	Code         []string          `kernel:"code"`
	Dev          []string          `kernel:"dev"`
	Ip           []string          `kernel:"ip"`
	Link         []string          `kernel:"link"`
	Sd           []string          `kernel:"sd"`
	Uart         []string          `kernel:"uart"`
	VGA          []string          `kernel:"vga"`
	Exports      map[string]string `kernel:"installs"`
}

func (k *Kernel) Hash() []byte {

	h := sha1.New()

	io.WriteString(h, k.Name)
	return h.Sum(nil)
}

func (k *Kernel) Build(c *build.Context) error {

	return nil
}

func (k *Kernel) Installs() map[string]string {
	return k.Exports
}

func (k *Kernel) GetName() string {
	return k.Name
}
func (k *Kernel) GetDependencies() []string {
	return k.Dependencies
}

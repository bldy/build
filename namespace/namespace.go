package namespace

import (
	"os"

	"github.com/pkg/errors"
)

const (
	MREPL int = iota
	MBEFORE
	MAFTER
)

type Namespace string

// Namespace is where builds are run, interface is the same as the plan9 namespaces
func Bind(new, old string, flags int)  {}
func Mount(new, old string, flags int) {}

func New(name string) error {
	if err := os.MkdirAll(name, 0755); err != nil {
		return errors.Wrap(err, "new namespace")
	}
	return nil
}

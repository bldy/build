package harvey

import "io"
import "bytes"

type Group struct {
	Name         string   `group:"name"`
	Dependencies []string `group:"deps"`
}

func (g *Group) Hash() []byte {

	return []byte{}
}

func (g *Group) Build() error {
	return nil
}

func (g *Group) GetName() string {
	return g.Name
}

func (g *Group) GetDependencies() []string {
	return g.Dependencies
}
func (g *Group) GetSource() string {
	return ""
}

func (g *Group) Reader() io.Reader {
	return bytes.NewBufferString("")
}

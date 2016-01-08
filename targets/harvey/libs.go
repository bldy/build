package harvey

import "sevki.org/build"

type Group struct {
	Name         string   `group:"name"`
	Dependencies []string `group:"deps"`
}

func (g *Group) Hash() []byte {

	return []byte(g.Name)
}

func (g *Group) Build(c *build.Context) error {

	return nil
}

func (g *Group) GetName() string {
	return g.Name
}

func (g *Group) GetDependencies() []string {
	return g.Dependencies
}
func (g *Group) Installs() map[string]string {
	return nil
}

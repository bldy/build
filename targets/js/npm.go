package js

import (
	"crypto/sha1"
	"fmt"

	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"sevki.org/build"
	"sevki.org/build/internal"
)

func init() {
	if err := internal.Register("npm_package", NpmPackage{}); err != nil {
		log.Fatal(err)
	}
}

type NpmPackage struct {
	Name         string   `npm_package:"name"`
	Version      string   `npm_package:"version"`
	Dependencies []string `npm_package:"deps"`
}

func (npm *NpmPackage) GetName() string {
	return npm.Name
}

func (npm *NpmPackage) GetDependencies() []string {
	return npm.Dependencies
}

func (npm *NpmPackage) Hash() []byte {
	h := sha1.New()
	io.WriteString(h, npm.Name)
	if npm.Version != "" {
		io.WriteString(h, npm.Version)
	} else {
		fmt.Fprintf(h, "%d", time.Now().UnixNano())
	}
	return h.Sum(nil)
}

func (npm *NpmPackage) Build(c *build.Context) error {
	if npm.Version == "" {
		return fmt.Errorf("NPM package %s failed to install, no version string")
	}

	params := []string{}
	params = append(params, "install")

	params = append(params, fmt.Sprintf("%s@%s", npm.Name, npm.Version))
	c.Println(strings.Join(append([]string{"npm"}, params...), " "))

	if err := c.Exec("npm", nil, params); err != nil {
		c.Println(err.Error())
		return fmt.Errorf(err.Error())
	}
	return nil
}

func (npm *NpmPackage) Installs() (installs map[string]string) {
	path := filepath.Join("node_modules", npm.Name)
	installs[path] = path
	return
}

package url

import (
	"fmt"
	"net/url"
	"path"

	"bldy.build/bldy/src/project"
)

type URL struct {
	url.URL
}

func Parse(s string) (*URL, error) {
	u, err := url.Parse(s)
	if u.Fragment == "" {
		_, u.Fragment = path.Split(u.Path)
	}
	if u.Scheme == "" {
		return Parse(fmt.Sprintf("file://%s/%s", project.RootKey, s))
	}
	return &URL{*u}, err
}
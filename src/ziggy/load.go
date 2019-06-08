package ziggy

import (
	"fmt"

	"bldy.build/bldy/src/build"
	"bldy.build/bldy/src/url"
)

type PackageLoader func(u *url.URL, bctx build.Context) (Package, error)

var loaders = make(map[string]PackageLoader)

func Register(scheme string, loader PackageLoader) {
	loaders[scheme] = loader
}

/*
func init() {
	loaders["file"] = fileLoader
	loaders["http"] = httpLoader
	loaders["https"] = httpLoader
}
*/
func Load(u *url.URL, bctx build.Context) (Package, error) {
	load, ok := loaders[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("%q is not a supported loader protocol", u.Scheme)
	}

	return load(u, bctx)
}

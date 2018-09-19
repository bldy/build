package builder

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"bldy.build/build"
	"bldy.build/build/graph"
	"bldy.build/build/namespace"
	"bldy.build/build/namespace/docker"
	"bldy.build/build/namespace/host"
)

var (
	ErrHostNotAvailable = errors.New("this compilation target is not compatible to run on this plan")
)

func nodeid(n *graph.Node) string {
	return fmt.Sprintf("%s-%s-bldy-%s-%x", n.Target.Name(), runtime.GOARCH, runtime.GOOS, n.HashNode())
}
func (b *Builder) newnamespace(n *graph.Node) (namespace.Namespace, error) {
	switch {
	case n.Target.Platform() == build.HostPlatform:
		return host.New(filepath.Join(*b.config.Cache, nodeid(n)))
	case n.Target.Platform().Repo() == "docker":
		return docker.New(n.Target.Platform(), filepath.Join(*b.config.Cache, nodeid(n)), nodeid(n), "debian:jessie")
	}
	return nil, ErrHostNotAvailable
}

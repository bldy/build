package docker

import (
	"context"
	"fmt"

	"bldy.build/build/label"
	"bldy.build/build/namespace"
	"bldy.build/build/namespace/host"
	"github.com/ahmetalpbalkan/dexec"
	docker "github.com/fsouza/go-dockerclient"
	"sevki.org/x/debug"
)

var d dexec.Docker

func init() {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err == nil {
		d = dexec.Docker{client}
	} else {
		panic(err)
	}

}

type Namespace struct {
	namespace.Namespace

	e      dexec.Execution
	opts   *docker.Config
	mounts []string
}

func New(l label.Label, cachedir, id, image string) (namespace.Namespace, error) {
	ns, err := host.New(cachedir)
	opts := &docker.Config{
		Image:  image,
		Mounts: []docker.Mount{},
	}
	if err != nil {
		return nil, err
	}
	return &Namespace{ns, nil, opts, []string{}}, nil
}
func (n *Namespace) Bind(new string, old string, flags int) {
	debug.Println(new, old)
	n.opts.Mounts = append(n.opts.Mounts, docker.Mount{
		Name:        "w00t",
		Source:      old,
		Destination: new,
		RW:          false,
		Driver:      "local",
	})
}

func (n *Namespace) Mount(new string, old string, flags int) {
	panic("not implemented")
}

func (n *Namespace) MountWorkspace(s string) {
	n.mounts = append(n.mounts, s) 

}

func (n *Namespace) Cmd(ctx context.Context, cmd string, args ...string) namespace.Cmd {
	binds := []string{}
	for _, s := range n.mounts {
		binds = append(binds, fmt.Sprintf("%s:%s", s, s))
	}

	if n.e == nil {
		n.e, _ = dexec.ByCreatingContainer(docker.CreateContainerOptions{
			Config:     n.opts,
			HostConfig: &docker.HostConfig{Binds: binds},
		})
	}
	debug.Println(binds)

	return d.Command(n.e, cmd, args...)
}

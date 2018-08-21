package deb

import (
	"compress/gzip"
	"fmt"
	"log"
	"net/http"
	"path"

	"bldy.build/build/executor"
	"bldy.build/build/label"
	"bldy.build/build/racy"
	"github.com/pkg/errors"
)

// A Debian repository contains several releases.
// Debian releases are named after characters from the "Toy Story" movies (wheezy, jessie, stretch, ...).
// The codenames have aliases, so called suites (stable, oldstable, testing, unstable).
// A release is divided in several components.
// In Debian these are named main, contrib, and non-free and indicate the licensing terms of the software they contain.
// A release also has packages for various architectures (amd64, i386, mips, powerpc, s390x, ...) as well as sources and architecture independent packages.
type Rule struct {
	name string

	Repo      string
	Release   string
	Component string
	Arch      string
	Version   string
	Pkg       string

	Deps []label.Label
	hash []byte
	Outs []string
}

func NewDebianRule(
	name string,

	repo string,
	release string,
	component string,
	arch string,
	version string,
	pkg string,

	hash []byte,
	outputs []string,
) *Rule {
	return &Rule{
		name:      name,
		hash:      hash,
		Repo:      repo,
		Release:   release,
		Component: component,
		Pkg:       pkg,
		Arch:      arch,
		Version:   version,
		Outs:      outputs,
	}
}

func (r *Rule) Name() string {
	return r.name
}

func (r *Rule) Dependencies() []label.Label {
	return r.Deps
}

func (r *Rule) Outputs() []string {
	return r.Outs
}

func (r *Rule) Hash() []byte {
	h := racy.New()
	h.Write(r.hash)
	h.HashStrings(
		r.Pkg, r.Arch, r.Version,
	)
	return h.Sum(nil)
}

func (r *Rule) Build(e *executor.Executor) error {
	// https://deb.debian.org/debian/dists/stable/main/binary-amd64/Packages.gz
	repopkgsurl := "https://" + path.Join(r.Repo, "dists", r.Release, r.Component, fmt.Sprintf("binary-%s", r.Arch), "Packages.gz")

	req, err := http.Get(repopkgsurl)
	if err != nil {
		log.Fatal(err)
	}
	gr, err := gzip.NewReader(req.Body)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("reading gzip %s errored", repopkgsurl))
	}
	repoScanner := NewRepoScanner(gr)
	found := false

	for repoScanner.Scan() {
		if repoScanner.ptr.Name == r.Pkg {
			if r.Version != "" && repoScanner.ptr.Version == r.Version {
				found = true
				break
			}
		}
	}
	req.Body.Close()
	if !found {
		return fmt.Errorf("could not find package %q in debian repo %q", r.Pkg, r.Repo)
	}
	pkgurl := "https://" + path.Join(r.Repo, repoScanner.ptr.Filename)
	req, err = http.Get(pkgurl)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("downloading deb %q failed", pkgurl))
	}
	pkgreader := NewPackageReader(req.Body, e)

	_, err = pkgreader.Read()

	return err
}

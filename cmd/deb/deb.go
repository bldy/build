// Copyright 2018 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deb // import "bldy.build/build/cmd/deb"
import (
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"

	"bldy.build/build/deb"
	"github.com/google/subcommands"
	"sevki.org/x/pretty"
)

type debCmdBase struct {
	repo      string
	release   string
	component string
	arch      string
	args      string
}

type DebRuleCmd struct {
	debCmdBase
}
type DebInspectCmd struct {
	debCmdBase
}

func (*DebRuleCmd) Name() string { return "deb-rule" }

func (*DebRuleCmd) Synopsis() string { return "prints a debian pkg as a bldy rule" }
func (*DebRuleCmd) Usage() string {
	return `deb-inspect foo
prints a debian pkg as a bldy rule
`
}
func (*DebInspectCmd) Name() string { return "deb-inspect" }

func (*DebInspectCmd) Synopsis() string { return "inspects a debian package" }
func (*DebInspectCmd) Usage() string {
	return `deb-inspect foo
Inspects a debian package for bldy related stuff
`
}

func (d *debCmdBase) SetFlags(f *flag.FlagSet) {
	f.StringVar(&d.repo, "repo", "cloudflaremirrors.com/debian", "apt repository")
	f.StringVar(&d.release, "release", "stable", "apt release")
	f.StringVar(&d.component, "component", "main", "apt component")
	f.StringVar(&d.arch, "arch", runtime.GOARCH, "cpu type")
}

func (d *debCmdBase) getpkg(name string) (*deb.Package, *deb.Deb, error) {

	repo := d.repo
	release := d.release
	component := d.component
	arch := d.arch

	// https://deb.debian.org/debian/dists/stable/main/binary-amd64/Packages.gz

	repopkgsurl := "https://" + path.Join(repo, "dists", release, component, fmt.Sprintf("binary-%s", arch), "Packages.gz")

	req, err := http.Get(repopkgsurl)
	if err != nil {
		return nil, nil, err
	}
	gr, err := gzip.NewReader(req.Body)
	if err != nil {
		return nil, nil, err
	}
	repoScanner := deb.NewRepoScanner(gr)
	found := false

	for repoScanner.Scan() {
		if repoScanner.Last().Name == name {
			found = true
			break
			/*	if r.Version != "" && repoScanner.ptr.Version == version {

				}*/
		}
	}
	req.Body.Close()
	if !found {

		return nil, nil, fmt.Errorf("could not find package %q in debian repo %q", name, repopkgsurl)
	}
	pkgurl := "https://" + path.Join(repo, repoScanner.Last().Filename)
	req, err = http.Get(pkgurl)
	if err != nil {
		fmt.Errorf("downloading deb %q failed", pkgurl)
		return nil, nil, fmt.Errorf("downloading deb %q failed", pkgurl)
	}
	pkgreader := deb.NewPackageReader(req.Body, nil)
	db, err := pkgreader.Read()
	if err != nil {
		return nil, nil, err
	}
	return repoScanner.Last(), db, nil
}

func (d *DebInspectCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if len(args) < 1 {
		return subcommands.ExitUsageError
	}
	name := args[0]
	ptr, _, err := d.getpkg(name)
	if err != nil {
		return subcommands.ExitFailure
	}
	l := map[string]interface{}{
		"Name":      name,
		"Arch":      d.arch,
		"Component": d.component,
		"Checksum":  fmt.Sprintf("%s:%x", "sha256", ptr.SHA256),
		"Pkg":       name,
		"Release":   d.release,
		"Repo":      d.repo,
		"Version":   ptr.Version,
		"Depends":   ptr.Depends,
	}
	fmt.Println(pretty.JSON(l))

	return subcommands.ExitSuccess
}

func (d *DebRuleCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if len(args) < 1 {
		return subcommands.ExitUsageError
	}
	name := args[0]

	ptr, db, err := d.getpkg(name)
	if err != nil {
		log.Println(err)
		return subcommands.ExitFailure
	}
	w := os.Stdout
	{
		io.WriteString(w, "deb(\n")
		for k, v := range map[string]string{
			"name":      name,
			"arch":      ptr.Arch,
			"checksum":  fmt.Sprintf("%s:%x", "sha256", ptr.SHA256),
			"component": d.component,
			"pkg":       name,
			"release":   d.release,
			"repo":      d.repo,
			"version":   ptr.Version,
		} {
			fmt.Fprintf(w, "   %s = %q,\n", k, v)
		}
		fmt.Fprintf(w, "   outputs = [\n")
		for _, f := range db.Files {
			fmt.Fprintf(w, "      %q,\n", f[2:])
		}
		io.WriteString(w, "   ]\n")
		io.WriteString(w, ")\n")
	}

	return subcommands.ExitSuccess
}

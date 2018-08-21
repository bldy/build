package deb

import (
	"log"
	"net/http"
	"testing"
)

func TestReadPkg(t *testing.T) {
	// control in one of these packages has an extra byte
	for _, pkg := range []string{
		"https://cdn-aws.deb.debian.org/debian/pool/main/l/lua5.1/lua5.1_5.1.5-8.1+b2_amd64.deb",
		"https://cdn-aws.deb.debian.org/debian/pool/main/l/luacheck/lua-check_0.17.1-1_all.deb",
	} {
		req, err := http.Get(pkg)
		if err != nil {
			log.Fatal(err)
		}
		pkgReader := NewPackageReader(req.Body, nil)
		if _, err := pkgReader.Read(); err != nil {
			t.Log(err)
			t.Fail()
		}
	}
}

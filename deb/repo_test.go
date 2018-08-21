package deb

import (
	"compress/gzip"
	"log"
	"net/http"
	"testing"
)

func TestReadRepoFromRemote(t *testing.T) {
	req, err := http.Get("https://cdn-aws.deb.debian.org/debian/dists/stable/main/binary-amd64/Packages.gz")
	if err != nil {
		log.Fatal(err)
	}
	r, err := gzip.NewReader(req.Body)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	repo := NewRepoScanner(r)

	for i := 0; i < 10; i++ {
		repo.Scan()
	}
}

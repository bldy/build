// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"sevki.org/build/builder"
	"sevki.org/lib/prettyprint"
)

var (
	targetName string
)

func server(t string) {
	targetName = t

	http.HandleFunc("/static/", static)
	http.HandleFunc("/graph/", graph)
	http.HandleFunc("/", index)
	log.Fatal(http.ListenAndServe(":8081", nil))

}
func index(w http.ResponseWriter, r *http.Request) {
	wd := "/Users/sevki/Code/go/src/sevki.org/build"
	f, err := os.Open(filepath.Join(wd, "graph/index.html"))
	if err != nil {
		http.Error(w, err.Error()+":\n"+filepath.Join(wd, "graph/index.html"), http.StatusNotFound)
	}
	io.Copy(w, f)
}
func static(w http.ResponseWriter, r *http.Request) {
	wd := "/Users/sevki/Code/go/src/sevki.org/build"
	f, err := os.Open(filepath.Join(wd, "graph", r.URL.Path[1:]))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	io.Copy(w, f)
}
func graph(w http.ResponseWriter, r *http.Request) {
	c := builder.New()

	if c.ProjectPath == "" {
		fmt.Fprintf(os.Stderr, "You need to be in a git project.\n\n")
		printUsage()
	}
	c.Parse(targetName)

	w.Write([]byte(prettyprint.AsJSON(c.Root)))

}

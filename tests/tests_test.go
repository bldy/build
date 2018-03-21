package tests

import (
	"context"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"bldy.build/bldy/tap"
	"bldy.build/build/builder"
	"bldy.build/build/graph"
)

var tests = []struct {
	name  string
	label string
}{
	{
		name:  "empty",
		label: "//empty:nothing",
	},
	{
		name:  "run",
		label: "//run:sh",
	},
}

func setup(t *testing.T) string {
	wd := path.Join(build.Default.GOPATH, "src", "bldy.build", "build", "tests", "testdata")
	os.Chdir(wd)
	return wd
}

func TestGraph(t *testing.T) {
	wd := setup(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g, err := graph.New(wd, test.label)
			if err != nil {
				t.Fatal(err)
			}
			if g == nil {
				t.Fail()
			}
		})
	}
}

func TestBuild(t *testing.T) {
	wd := setup(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g, err := graph.New(wd, test.label)
			if err != nil {
				t.Fatal(err)
			}
			if g == nil {
				t.Fail()
			}
			tmpDir, _ := ioutil.TempDir("", fmt.Sprintf("bldy_test_%s_", test.name))
			b := builder.New(g, &builder.Config{
				UseCache: false,
				BuildOut: &tmpDir,
			})
			cpus := 1
			ctx := context.Background()
			go b.Execute(ctx, cpus)
			display := tap.New()

			go display.Display(b.Updates, cpus)
		})
	}
}

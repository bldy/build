package integration

import (
	"context"
	"fmt"
	gobuild "go/build"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"bldy.build/build"
	"sevki.org/x/mock"

	"bldy.build/build/builder"
	"bldy.build/build/graph"
	"bldy.build/build/integration/internal"
	"bldy.build/build/label"
)

var tests = []struct {
	name  string
	label string
	err   error
}{
	{
		name:  "empty",
		label: "//empty:nothing",
		err:   nil,
	},
	{
		name:  "hello",
		label: "//hello:world",
		err:   nil,
	},
}

func setup(t *testing.T) string {
	wd := path.Join(gobuild.Default.GOPATH, "src", "bldy.build", "build", "integration", "testdata")
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

			tn := internal.NewNotifier(t)
			mock := tn.(mock.Mock)

			lbl, _ := label.Parse(test.label)
			mock.EXPECT().(builder.Notifier).Notify(build.Success, &graph.Node{Label: *lbl})

			b := builder.New(
				g,
				&builder.Config{
					UseCache: false,
					BuildOut: &tmpDir,
				},
				tn,
			)
			cpus := 1
			ctx := context.Background()
			b.Execute(ctx, cpus)
		})
	}
}

package integration

import (
	"context"
	"fmt"
	gobuild "go/build"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"
	"time"

	"bldy.build/build"
	"bldy.build/build/builder"
	"bldy.build/build/graph"
	"sevki.org/x/debug"
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
		name:  "binary",
		label: "//cc:hello",
		err:   nil,
	},
}

func setup(t *testing.T) string {
	wd := path.Join(gobuild.Default.GOPATH, "src", "bldy.build", "build", "tests", "testdata")
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

type testNotifier struct {
	t *testing.T
}

func (t *testNotifier) Update(n *graph.Node) {
	switch n.Status {
	case build.Building:
		t.t.Logf("Started building %s ", n.Label.String())
	default:
		t.t.Logf("Started %d %s ", n.Status, n.Label.String())

	}

}

func (t *testNotifier) Error(err error) {
	t.t.Fail()
	t.t.Logf("Errored:%+v\n", err)
}

func (t *testNotifier) Done(d time.Duration) {
	t.t.Logf("Finished building in %s\n", d)

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

			b := builder.New(
				g,
				&builder.Config{
					UseCache: false,
					BuildOut: &tmpDir,
				},
				&testNotifier{t},
			)
			cpus := 1
			ctx := context.Background()
			b.Execute(ctx, cpus)

			files, err := ioutil.ReadDir(tmpDir)

			if err != nil {
				log.Fatal(err)
			}
			for _, file := range files {
				debug.Println(file.Name())
			}
		})
	}
}

package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"
	"time"

	"bldy.build/bldy/src/build"
	"bldy.build/bldy/src/builder"
	"bldy.build/bldy/src/graph"
	"bldy.build/bldy/src/url"
)

var wd = func() string { wd, _ := os.Getwd(); return wd }()
var tests = []struct {
	name string
	url  string
	wd   string
	err  error
}{

	{
		name: "run",
		url:  "rust#rust-bin",
		wd:   path.Join(wd, "testdata"),
		err:  nil,
	},
}

type testNotifier struct {
	t *testing.T
}

func (t *testNotifier) Update(n *graph.Node) {
	switch n.Status {
	case build.Building:
		t.t.Logf("Started building %s ", n.ID)
	default:
		t.t.Logf("Started %d %s ", n.Status, n.ID)

	}

}

func (t *testNotifier) Error(err error) {
	t.t.Fail()
	t.t.Logf("error: %+v\n", err)
}

func (t *testNotifier) Done(d time.Duration) {
	t.t.Logf("Finished building in %s\n", d)

}

func TestBuild(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			u, _ := url.Parse(test.url)
			g, err := graph.New(u, test.wd)
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
					Fresh:    true,
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
				_ = file
			}
		})
	}
}

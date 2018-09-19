package skylark

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bldy.build/build/label"
	"bldy.build/build/workspace"
	fuzz "github.com/google/gofuzz"
	"github.com/icrowley/fake"
)

func TestFile(t *testing.T) {

	// lets create a tmp dir
	dir, _ := ioutil.TempDir("", "bldy_fuzz")
	// and make it a workspace
	os.Create(filepath.Join(dir, "WORKSPACE"))

	bytz, err := ioutil.ReadFile("testdata/hashtest/hashtest.sky")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	ioutil.WriteFile(filepath.Join(dir, "hashtest.sky"), bytz, 0755)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	tmpl := template.Must(template.New("BUILD").ParseFiles("testdata/hashtest/BUILD.tmpl"))
	f, err := os.Create(filepath.Join(dir, "BUILD"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	data := randomData(dir)

	if err := tmpl.Execute(f, data); err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	ws, err := workspace.New(dir)
	vm, _ := New(ws)
	_ = vm
	x := make(map[string]string)
	for _, d := range data {
		l, err := label.Parse(fmt.Sprintf("//.:%s", d.Name))
		if err != nil {
			panic(err)
		}

		target, err := vm.GetTarget(l)
		if err != nil {
			t.Log(err)
			t.Fail()
			return
		}
		h := target.Hash()
		hs := fmt.Sprintf("%x", h)

		if e, ok := x[hs]; ok && e != hs {
			t.Logf("%s and %s have the same hash %x", l.String(), e, hs)
			t.FailNow()
		}

		x[hs] = l.String()
		break
	}

	os.RemoveAll(dir)

}

type testData struct {
	Name  string
	Files []string
}

func randomData(wd string) []testData {

	tds := []testData{}
	for i := 0; i < 10+r(10); i++ {
		td := testData{
			Name: fake.WordsN(1),
		}
		for k := 0; k < 1+r(10); k++ {
			n := fake.WordsN(1)
			f, _ := os.Create(filepath.Join(wd, fmt.Sprintf("%s.c", n)))
			fzz := fuzz.New()
			var s string
			for j := 0; j < 10+r(50); j++ {
				fzz.Fuzz(&s)
				io.WriteString(f, s)
			}
			td.Files = append(td.Files, n)
		}
		tds = append(tds, td)
	}
	return tds
}

func r(i int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(i)
}
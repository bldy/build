package url

import (
	"testing"

	"sevki.org/x/pretty"
)

func parseURL(s string) URL {
	u, _ := Parse(s)
	return *u
}

var urls = []struct {
	in       string
	host     string
	fragment string
}{
	{
		"src",
		"$PROJECTROOT",
		"src",
	},
	{
		"src#libc",
		"$PROJECTROOT",
		"libc",
	},
	{
		"http://bldy.build/bldy/src#libc",
		"bldy.build",
		"libc",
	},
}

func TestNewURL(t *testing.T) {
	for _, test := range urls {
		t.Run(test.fragment, func(t *testing.T) {
			u, _ := Parse(test.in)
			if test.host != u.Host {
				t.Logf("expected %q got %q instead", test.host, u.Host)
				t.Log(pretty.JSON(u))
				t.Fail()
			}
			if test.fragment != u.Fragment {
				t.Logf("expected %q got %q instead", test.fragment, u.Fragment)
				t.Log(pretty.JSON(u))
				t.Fail()
			}
		})
	}
}

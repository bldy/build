package racy

import (
	"bytes"
	"testing"
)

func TestXOR(t *testing.T) {
	tests := []struct {
		name string
		a, b []byte
	}{
		{
			a: []byte{},
			b: hashString("a"),
		},
		{
			a: hashString("a"),
			b: []byte{},
		},
		{
			a: hashString("a"),
			b: hashString("a"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			xord := XOR(test.a, test.b)

			if bytes.Compare(xord, []byte{}) == 0 {
				t.Logf("\"%x\" and \"%x\" are the same", xord, []byte{})
				t.Fail()
			}
		})
	}
}

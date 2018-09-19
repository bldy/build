package racy

import (
	"bytes"
)

// XORBytes takes two byte slices and XORs them
func XORBytes(a, b []byte) []byte {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dst []byte
	for i := 0; i < n; i++ {
		dst = append(dst, a[i]^b[i])
	}
	return dst
}

// XOR takes 2 or more byte slices and xors them.
func XOR(hs ...[]byte) []byte {
	if len(hs) == 1 {
		return hs[0]
	}

	var dst []byte

	for _, h := range hs {
		if len(h) == 0 {
			continue
		}
		if len(dst) == 0 {
			dst = h
			continue
		}
		a := dst
		b := h
		// this is a stupid hack and I'm ashamed of it
		if bytes.Compare(a, b) == 0 {
			a = hashString(string(a))
		}
		dst = XORBytes(a, b)
	}
	return dst
}

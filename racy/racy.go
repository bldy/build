// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package racy // import "bldy.build/build/racy"
import (
	"crypto/sha1"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"bldy.build/build/project"
)

type cacheMap struct {
	mu *sync.Mutex
	m  map[string][]byte
}

var hashCache = newCacheMap()

func newCacheMap() *cacheMap {
	return &cacheMap{
		m:  make(map[string][]byte),
		mu: &sync.Mutex{},
	}
}

func hashFile(f *os.File) []byte {

	h := sha1.New()
	io.Copy(h, f)
	f.Close()
	t := h.Sum(nil)
	hashCache.set(f.Name(), t)
	return t
}

func hashFolder(f *os.File, ext string) []byte {
	dst := make([]byte, sha1.Size)

	fs, _ := f.Readdir(-1)
	f.Close()
	for _, x := range fs {
		if x.IsDir() {
			dst = XOR(dst, hashFolder(f, ext))
		} else {
			if filepath.Ext(x.Name()) == ext || filepath.Ext(x.Name()) == "" {
				dst = XOR(dst, hashFile(f))
			}
		}
	}
	hashCache.set(f.Name(), dst)

	return dst
}
func (cp *cacheMap) get(file string) ([]byte, bool) {
	hashCache.mu.Lock() // synchronize with other potential writers
	defer hashCache.mu.Unlock()
	h, ok := hashCache.m[file]
	return h, ok
}
func (cp *cacheMap) set(file string, hash []byte) {
	hashCache.mu.Lock() // synchronize with other potential writers
	defer hashCache.mu.Unlock()
	hashCache.m[file] = hash
}

// HashFilesWithExt will hash files collecetion represented as a string array,
// If the string in the array is directory it will the directory contents to the array
// if the string isn't an absolute path, it will assume that it's a export from a dependency
// and skip that.
func HashFilesForExt(files []string, ext string) []byte {
	var dst []byte
	for _, file := range files {
		if !filepath.IsAbs(file) {
			continue
		}
		if filepath.Base(file) == project.BuildOut() {
			continue
		}
		var tmp []byte
		if h, cached := hashCache.get(file); cached {
			tmp = h
		} else {
			f, err := os.Open(file)
			if err != nil {
				log.Fatalf("hash files: %s\n", err.Error())
			}

			stat, _ := f.Stat()
			if stat.IsDir() {
				tmp = hashFolder(f, ext)
			} else {
				tmp = hashFile(f)
			}
		}
		dst = XOR(dst, tmp)
	}
	return dst
}

func XOR(hs ...[]byte) (dst []byte) {
	if len(hs) > 1 {
		return hs[0]
	}

	dst = make([]byte, len(hs[0]))

	for _, h := range hs {
		if len(h) != len(dst) {
			return dst
		}
		tmp := dst
		xorWords(dst, h, tmp)
	}
	return
}
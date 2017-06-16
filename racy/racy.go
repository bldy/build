// Copyright 2017 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package racy // import "bldy.build/build/racy"
import (
	"crypto/sha1"
	"crypto/sha256"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"bldy.build/build"
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
		if filepath.Ext(file) != ext || filepath.Ext(file) == "" {
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
		if len(dst) == 0 {
			dst = tmp
		} else {
			dst = XOR(dst, tmp)
		}
	}

	return dst
}

func hashFiles(files []string, exts []string) []byte {

	var dst []byte
	for _, file := range files {
		if !filepath.IsAbs(file) {
			continue
		}
		if filepath.Base(file) == project.BuildOut() {
			continue
		}
		stat, err := os.Stat(file)
		if err != nil {
			log.Fatalf("stating file %q: %s\n", file, err.Error())
		}

		fileExt := filepath.Ext(file)
		// Is there a extension filter?
		if !stat.IsDir() && len(exts) > 0 {
			validExtension := false
			for _, ext := range exts {
				if fileExt == ext {
					validExtension = true
				}
			}
			if !validExtension {
				continue
			}
		}

		var tmp []byte
		if h, cached := hashCache.get(file); cached {
			tmp = h
		} else {
			f, err := os.Open(file)
			if err != nil {
				log.Fatalf("hash files: %s\n", err.Error())
			}
			if stat.IsDir() {
				fs, _ := f.Readdirnames(-1)
				f.Close()
				for i, s := range fs {
					fs[i] = path.Join(file, s)
				}

				tmp = hashFiles(fs, exts)
			} else {
				tmp = hashFile(f)
			}

		}
		if len(dst) == 0 {
			dst = tmp
		} else {
			dst = XOR(dst, tmp)
		}
	}

	return dst
}

func HashTarget(target build.Target) []byte {
	var dst []byte
	typ := reflect.TypeOf(target).Elem()
	val := reflect.ValueOf(target).Elem()

	if typ.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		dst = XOR(dst, hashString(field.PkgPath))
		dst = XOR(dst, hashString(field.Name))

		tag := field.Tag.Get("build")
		if !(tag == "path" || tag == "expand") {
			continue
		}
		extTag := field.Tag.Get("ext")
		splitTags := []string{}
		if extTag != "" {
			splitTags = strings.Split(extTag, ",")
		}
		dst = XOR(dst, hashPath(val.Field(i), splitTags))
	}

	dst = XOR(dst, hashValue(val))

	return dst
}
func hashPath(v reflect.Value, exts []string) []byte {
	var dst []byte

	switch v.Kind() {
	case reflect.String:
		file := v.String()
		dst = XOR(dst, hashFiles([]string{file}, exts))
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			dst = XOR(dst, hashPath(v.Index(i), exts))
		}
	}
	return dst
}
func hashValue(v reflect.Value) []byte {
	var dst []byte

	switch v.Kind() {
	case reflect.String:
		dst = XOR(dst, hashString(v.String()))
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			dst = XOR(dst, hashValue(f))
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			dst = XOR(dst, hashValue(v.Index(i)))
		}
	}
	return dst
}
func hashString(s string) []byte {
	h := sha256.New()
	io.WriteString(h, s)
	return h.Sum(nil)
}
func HashStrings(h io.Writer, strs []string) {
	for _, str := range strs {
		io.WriteString(h, str)
	}
}

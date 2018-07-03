// Package racy deals with file cryptography
package racy

import (
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// NewHash returns a new hash.Hash
var NewHash = sha512.New

var (
	cacheMutex = new(sync.Mutex)
	hashCache  = make(map[string][]byte)
)

// Racy is used in hashing targets.
// https://www.kernel.org/pub/software/scm/git/docs/technical/racy-git.txt
type Racy struct {
	hash.Hash

	exts []string

	mu *sync.Mutex
	m  map[string][]byte
}

type Option func(*Racy)

// New takes nothing and returns a new Racy
func New(options ...Option) *Racy {
	x := &Racy{
		NewHash(),

		[]string{}, // by default we'll hash everything checkout `AllowExtension` option to limit the files hashed
		&sync.Mutex{},
		make(map[string][]byte),
	}
	for _, option := range options {
		option(x)
	}
	return x
}

func AllowExtension(ext string) Option {
	ext = strings.TrimLeft(ext, ".")
	return func(r *Racy) { r.exts = append(r.exts, ext) }
}

func hashString(s string) []byte {
	h := NewHash()
	io.WriteString(h, s)
	return h.Sum(nil)
}

func hashFile(file string) []byte {
CHECK:
	if sum, ok := hashCache[file]; ok {
		return sum
	}
	// For speediness, we'll cache the hash of the file
	// and hash that instead of opening the file,
	// reading it's contents everytime
	cacheMutex.Lock()
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("racy.hashFile: error opening file %q: %v", file, err)
	}
	tmpHash := NewHash()
	io.Copy(tmpHash, f)
	hashCache[file] = tmpHash.Sum(nil)
	f.Close()
	cacheMutex.Unlock()
	goto CHECK
}

func (r *Racy) hashDir(dir string) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", dir, err)
			return err
		}
		if info.IsDir() {
			r.hashDir(path)
		} else {
			r.hashFile(path)
		}
		return nil
	})
	log.Fatalf("racy.hashDir: error walking dir %q: %v", dir, err)
}

func (r *Racy) hashFile(file string) {
	if !r.allowedExt(file) {
		return
	}
	r.Write(hashFile(file))
}

func (r *Racy) allowedExt(file string) bool {
	ext := strings.TrimLeft(filepath.Ext(file), ".")
	for _, allowedExt := range r.exts {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

func (r *Racy) HashFiles(files ...string) {
	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			log.Fatalf("racy.HashFiles: error opening file: %v", err)
		}
		if !stat.IsDir() {
			r.hashFile(file)
		} else {
			r.hashDir(file)
		}
	}
	return
}

// HashStrings hashes strings written to Racy
func (r *Racy) HashStrings(strs ...string) {
	for _, str := range strs {
		io.WriteString(r, str)
	}
}

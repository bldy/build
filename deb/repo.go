package deb

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/mail"
	"strconv"
	"strings"

	"sevki.org/x/debug"
)

const description = "Description"

type stateFn func(*RepoScanner) stateFn

type RepoScanner struct {
	scanner *bufio.Scanner
	pkgs    []Package
	buf     string
	ptr     *Package
	state   stateFn
	count   int
}
type checksum []byte

func (c checksum) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%x"`, c)), nil
}

type Package struct {
	Name          string
	Priority      string
	Section       string
	Size          uint64
	InstalledSize uint64
	Maintainer    *mail.Address
	Vendor        *mail.Address
	Homepage      string
	Arch          string
	Source        string
	Version       string
	Replaces      string
	Provides      string
	Depends       []string
	Filename      string
	MD5           checksum
	SHA1          checksum
	SHA256        checksum
	SHA512        checksum
	License       string
	Description   *bytes.Buffer
}

func NewRepoScanner(r io.Reader) *RepoScanner {
	return &RepoScanner{scanner: bufio.NewScanner(r)}
}

func (r *RepoScanner) Last() *Package {
	return r.ptr
}
func (r *RepoScanner) Scan() bool {
	pkg := Package{}
	r.ptr = &pkg
	for r.state = scanLine; r.state != nil; {
		if !r.scanner.Scan() {
			return false
		}
		if err := r.scanner.Err(); err != nil {
			return false
		}
		r.buf = r.scanner.Text()
		if r.buf == "" {
			break
		}
		r.state = r.state(r)
	}
	r.pkgs = append(r.pkgs, pkg)
	return true
}

func splitKV(s string) (string, string) {
	i := strings.IndexByte(s, ':')
	if i <= 0 {
		return "", ""
	}
	return s[:i], strings.TrimSpace(s[i+1:])
}

func scanDescription(r *RepoScanner) stateFn {
	if len(r.buf) > 1 && r.buf[0] == ' ' {
		io.WriteString(r.ptr.Description, r.buf)
		return scanDescription
	}
	return scanLine
}

func scanLine(r *RepoScanner) stateFn {
	if len(r.buf) == 1 && r.buf[0] == '\n' {
		debug.Println("ASDASD")
		return nil
	}
	if len(r.buf) <= 0 {
		return nil
	}
	k, v := splitKV(r.buf)
	switch k {
	case "Depends":
		r.ptr.Depends = strings.Split(v, ", ")
	case "Package":
		r.ptr.Name = v
	case "Priority":
		r.ptr.Priority = v
	case "Section":
		r.ptr.Section = v
	case "Installed-Size":
		r.ptr.InstalledSize = shouldParseUInt64(v)
	case "Maintainer":
		r.ptr.Maintainer = shouldParseMailAddress(v)
	case "Vendor":
		r.ptr.Vendor = shouldParseMailAddress(v)
	case "Architecture":
		r.ptr.Arch = v
	case "Source":
		r.ptr.Source = v
	case "Version":
		r.ptr.Version = v
	case "Replaces":
		r.ptr.Replaces = v
	case "Provides":
		r.ptr.Provides = v
	case "Conflicts":
	// bldy is better at figuring this out
	case "License":
		r.ptr.License = v
	case "Homepage":
		r.ptr.Homepage = v
	case "Filename":
		r.ptr.Filename = v
	case "Size":
		r.ptr.Size = shouldParseUInt64(v)
	case "MD5sum":
		r.ptr.MD5 = shouldDecodeHex(v)
	case "SHA1":
		r.ptr.SHA1 = shouldDecodeHex(v)
	case "SHA256":
		r.ptr.SHA256 = shouldDecodeHex(v)
	case "SHA512":
		r.ptr.SHA512 = shouldDecodeHex(v)
	case "Description":
		r.ptr.Description = bytes.NewBufferString(v)
		return scanDescription

	default:
		// Don't care about these
		/*"Standards-Version",
		"Breaks",
		"Build-Ids",
		"Vcs-Browser",
		"Vcs-Git",
		"Built-Using",
		"Recommends",
		"Essential",
		"Suggests",
		"Multi-Arch",
		"Pre-Depends",
		"Auto-Built-Package":
		*/
	}

	return scanLine
}

// we don't much care about sizes, it's good to have these but not a deal breaker
// if a hex value comes here and we weren't prepared to handle it
func shouldParseUInt64(v string) uint64 {
	if u, err := strconv.ParseUint(v, 10, 64); err == nil {
		return u
	}
	return 0
}

func shouldParseMailAddress(v string) *mail.Address {
	if e, err := mail.ParseAddress(v); err == nil {
		return e
	}
	return nil
}

func shouldDecodeHex(v string) checksum {
	if decoded, err := hex.DecodeString(v); err == nil {
		return checksum(decoded)
	}
	return checksum([]byte{})
}

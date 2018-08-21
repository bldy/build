package deb

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"bldy.build/build/executor"
	"github.com/pkg/errors"
	"github.com/ulikunitz/xz"
	"sevki.org/x/debug"
)

const arHeader = "!<arch>\n"
const debianFileIdentifier = "debian-binary   "

type PackageReader struct {
	r *bufio.Reader
	e *executor.Executor
}

type header struct {
	Name    string
	ModTime time.Time
	Owner   int
	Group   int
	Mode    os.FileMode
	Size    int64
}

func NewPackageReader(r io.Reader, e *executor.Executor) *PackageReader {
	return &PackageReader{r: bufio.NewReader(r), e: e}
}
func (r *PackageReader) readArSignature() error {
	// archive file signature
	fileSignature := make([]byte, len(arHeader))
	if n, err := r.r.Read(fileSignature); err != nil || n != len(arHeader) || string(fileSignature) != arHeader {
		return errors.New("not a valid deb pkg, parsing header failed, wrong archive signature")
	}
	return nil
}

func (r *PackageReader) readHeader() (*header, error) {
	const headerSize = 60
	// read file identifier
	buf := make([]byte, headerSize)
	if n, err := r.r.Read(buf); err != nil || n != headerSize {
		return nil, errors.New("not a valid deb pkg, parsing header failed")
	}
	h := &header{}
	readBytesAsString := func(n int) string {
		var x string
		x, buf = string(buf[:n]), buf[n:]
		return x
	}
	h.Name = readBytesAsString(16)

	if unixTime, err := strconv.ParseInt(strings.TrimSpace(readBytesAsString(12)), 10, 64); err == nil {
		h.ModTime = time.Unix(unixTime, 0)
	} else {
		return nil, errors.Wrap(err, "cannot convert size to int, please see doc")
	}

	var err error
	if h.Owner, err = strconv.Atoi(strings.TrimSpace(readBytesAsString(6))); err != nil {
		return nil, errors.Wrap(err, "cannot parse owner")
	}
	if h.Group, err = strconv.Atoi(strings.TrimSpace(readBytesAsString(6))); err != nil {
		return nil, errors.Wrap(err, "cannot parse group")
	}

	if mode, err := strconv.ParseUint(strings.TrimSpace(readBytesAsString(8)), 10, 32); err == nil {
		h.Mode = os.FileMode(uint32(mode))
	} else {
		return nil, errors.Wrap(err, "cannot convert size to int, please see doc")
	}

	if h.Size, err = strconv.ParseInt(strings.TrimSpace(readBytesAsString(10)), 10, 64); err != nil {
		return nil, errors.Wrap(err, "cannot convert size to int, please see doc")
	}

	endChar := readBytesAsString(2)
	if endChar != string([]byte{'`', '\n'}) {
		return nil, errors.New("corrupted package")
	}

	return h, nil
}

func parseDebSig(*PackageReader, *header) error {
	return nil
}

// follows https://upload.wikimedia.org/wikipedia/commons/6/67/Deb_File_Structure.svg
func (r *PackageReader) Read() (*Deb, error) {
	if err := r.readArSignature(); err != nil {
		return nil, err
	}

	d := &Deb{}
	for _, part := range []struct {
		prefix string
		read   fileParser
	}{
		{prefix: "debian", read: d.parseDebVersion},
		{prefix: "control", read: d.readArchive},
		{prefix: "data", read: d.installArchive(r.e)},
	} {
		header, err := r.readHeader()
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(header.Name, part.prefix) {
			if err := part.read(r.r, header); err != nil {
				return nil, err
			}
		}
	}
	return d, nil
}

type Deb struct {
	Version string
	Files   []string
}

func (d *Deb) parseDebVersion(r *bufio.Reader, h *header) error {
	buf := make([]byte, h.Size)
	if n, err := r.Read(buf); err != nil || int64(n) != h.Size {
		return errors.New("not a valid deb pkg, parsing header failed")
	}
	d.Version = string(buf)
	return nil
}

type createFileFunc func(name string) (*os.File, error)
type fileParser func(*bufio.Reader, *header) error

func (d *Deb) installArchive(e *executor.Executor) fileParser {
	return func(bufr *bufio.Reader, h *header) error {
		var err error
		var r io.Reader
		r = io.LimitReader(bufr, (h.Size))
		if r, err = compressionReader(r, h); err != nil {
			return err
		}
		tr := tar.NewReader(r)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break // End of archive
			}
			if err != nil {
				return err
			}
			if e == nil {
				if hdr.Typeflag == tar.TypeReg {
					d.Files = append(d.Files, hdr.Name)
				}
				if _, err := io.CopyN(ioutil.Discard, tr, hdr.Size); err != nil {
					return err
				}
				continue
			}
			switch hdr.Typeflag {
			case tar.TypeReg:
				d.Files = append(d.Files, hdr.Name)

				f, err := e.OpenFile(hdr.Name, os.O_RDWR|os.O_CREATE, os.ModePerm)
				if err != nil {
					panic(err)
					return err
				}
				if _, err := io.CopyN(f, tr, hdr.Size); err != nil {
					return err
				}
				f.Close()
			case tar.TypeDir:
				e.Mkdir(hdr.Name)
			}
		}
		return nil
	}
}

func (d *Deb) readArchive(bufr *bufio.Reader, h *header) error {
	var err error
	var r io.Reader
	r = io.LimitReader(bufr, (h.Size + h.Size%2))
	if r, err = compressionReader(r, h); err != nil {
		return err
	}
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		if _, err := io.CopyN(ioutil.Discard, tr, hdr.Size); err != nil {
			return err
		}
	}
	return nil
}

func compressionReader(r io.Reader, h *header) (io.Reader, error) {
	switch path.Ext(strings.TrimSpace(h.Name)) {
	case ".gz":
		return gzip.NewReader(r)
	case ".xz":
		return xz.NewReader(r)
	default:
		debug.Printf("%q", path.Ext(h.Name))
	}
	return r, nil
}
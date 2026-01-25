package tarfs

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"

	"github.com/unstoppablemango/ihfs"
)

type Cache interface {
	ihfs.WriteFile
	ihfs.Mkdir
}

type File struct {
	hdr *tar.Header
	buf bytes.Buffer
}

// Close implements [fs.File].
func (f *File) Close() error {
	return nil
}

// Read implements [fs.File].
func (f *File) Read(b []byte) (int, error) {
	return f.buf.Read(b)
}

// Stat implements [fs.File].
func (f *File) Stat() (fs.FileInfo, error) {
	return f.hdr.FileInfo(), nil
}

func (f *File) Name() string {
	return f.hdr.Name
}

func (f *File) readFrom(r io.Reader) error {
	n, err := f.buf.ReadFrom(r)
	if err != nil {
		return err
	}
	if n == f.hdr.Size {
		return nil
	}

	return SizeError{
		Expected: f.hdr.Size,
		Actual:   n,
	}
}

type SizeError struct {
	Expected int64
	Actual   int64
}

func (err SizeError) Error() string {
	return fmt.Sprintf("expected=%d actual=%d", err.Expected, err.Actual)
}

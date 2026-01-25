package tarfs

import (
	"archive/tar"
	"bytes"
	"io/fs"
)

type File struct {
	hdr *tar.Header
	buf *bytes.Buffer
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

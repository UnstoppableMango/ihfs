package tarfs

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
)

// File represents a file in a tar archive.
type File struct {
	io.Reader
	hdr *tar.Header
}

// Close implements [fs.File].
func (f *File) Close() error {
	return nil
}

// Stat implements [fs.File].
func (f *File) Stat() (fs.FileInfo, error) {
	return f.hdr.FileInfo(), nil
}

// Name returns the name of the tar entry.
func (f *File) Name() string {
	return f.hdr.Name
}

type fileData struct {
	hdr  *tar.Header
	data []byte
}

func (fd fileData) file() *File {
	return &File{
		hdr:    fd.hdr,
		Reader: bytes.NewReader(fd.data),
	}
}

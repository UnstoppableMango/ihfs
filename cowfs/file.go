package cowfs

import (
	"io"
	"io/fs"

	"github.com/unstoppablemango/ihfs"
)

type File struct {
	name  string
	base  ihfs.File
	layer ihfs.File
}

// Close implements [fs.File].
func (f *File) Close() error {
	// Base should be closed first so that the overlay has a newer
	// timestamp, otherwise the cache will never get hit.
	if f.base != nil {
		f.base.Close()
	}
	if f.layer != nil {
		return f.layer.Close()
	}

	return BADFD
}

// Read implements [fs.File].
func (f *File) Read(b []byte) (int, error) {
	if f.layer != nil {
		n, err := f.layer.Read(b)
		if (err == nil || err == io.EOF) && f.base != nil {
			// if _, seekErr := f.base.
		}
	}
	if f.base != nil {
		return f.base.Read(b)
	}

	return 0, BADFD
}

// Stat implements [fs.File].
func (f *File) Stat() (fs.FileInfo, error) {
	panic("unimplemented")
}

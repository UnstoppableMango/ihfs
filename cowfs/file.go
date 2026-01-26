package cowfs

import (
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
	return nil
}

// Read implements [fs.File].
func (f *File) Read([]byte) (int, error) {
	panic("unimplemented")
}

// Stat implements [fs.File].
func (f *File) Stat() (fs.FileInfo, error) {
	panic("unimplemented")
}

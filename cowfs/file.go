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
	panic("unimplemented")
}

// Read implements [fs.File].
func (f *File) Read([]byte) (int, error) {
	panic("unimplemented")
}

// Stat implements [fs.File].
func (f *File) Stat() (fs.FileInfo, error) {
	panic("unimplemented")
}

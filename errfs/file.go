package errfs

import (
	"github.com/unstoppablemango/ihfs"
)

// File is a file that returns a fixed error for all operations.
type File struct {
	err error
}

// NewFile creates a new [File] that returns err for all operations.
func NewFile(err error) *File {
	return &File{err: err}
}

// Close implements [ihfs.File].
func (f *File) Close() error {
	return f.err
}

// Read implements [ihfs.File].
func (f *File) Read(_ []byte) (int, error) {
	return 0, f.err
}

// Stat implements [ihfs.File].
func (f *File) Stat() (ihfs.FileInfo, error) {
	return nil, f.err
}

// ReadDir implements [ihfs.ReadDirFile].
func (f *File) ReadDir(_ int) ([]ihfs.DirEntry, error) {
	return nil, f.err
}

// ReadDirNames implements [ihfs.DirNameReader].
func (f *File) ReadDirNames(_ int) ([]string, error) {
	return nil, f.err
}

// Seek implements [io.Seeker].
func (f *File) Seek(_ int64, _ int) (int64, error) {
	return 0, f.err
}

// Write implements [io.Writer].
func (f *File) Write(_ []byte) (int, error) {
	return 0, f.err
}

// ReadAt implements [io.ReaderAt].
func (f *File) ReadAt(_ []byte, _ int64) (int, error) {
	return 0, f.err
}

// WriteAt implements [io.WriterAt].
func (f *File) WriteAt(_ []byte, _ int64) (int, error) {
	return 0, f.err
}

// WriteString implements [io.StringWriter].
func (f *File) WriteString(_ string) (int, error) {
	return 0, f.err
}

// Sync implements [ihfs.Syncer].
func (f *File) Sync() error {
	return f.err
}

// Truncate implements [ihfs.Truncater].
func (f *File) Truncate(_ int64) error {
	return f.err
}

package ghfs

import (
	"io"
	"io/fs"
	"path"
	"time"
)

// Dir is a virtual directory in the GitHub filesystem.
// It implements [fs.ReadDirFile] and [fs.FileInfo] with no entries.
type Dir struct {
	name string
}

func (d *Dir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.name, Err: fs.ErrInvalid}
}

func (d *Dir) Close() error { return nil }

func (d *Dir) Stat() (fs.FileInfo, error) { return d, nil }

func (d *Dir) ReadDir(n int) ([]fs.DirEntry, error) {
	if n > 0 {
		return nil, io.EOF
	}
	return nil, nil
}

func (d *Dir) Name() string       { return path.Base(d.name) }
func (d *Dir) IsDir() bool        { return true }
func (d *Dir) Mode() fs.FileMode  { return fs.ModeDir | 0555 }
func (d *Dir) ModTime() time.Time { return time.Time{} }
func (d *Dir) Size() int64        { return 0 }
func (d *Dir) Sys() any           { return nil }

package ghfs

import (
	"io"
	"io/fs"
	"path"
)

// Dir is a virtual directory in the GitHub filesystem.
// It implements [fs.ReadDirFile] with no entries.
type Dir struct {
	name string
}

func (d *Dir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.name, Err: fs.ErrInvalid}
}

func (d *Dir) Close() error { return nil }

func (d *Dir) Stat() (fs.FileInfo, error) {
	return &FileInfo{name: path.Base(d.name), isDir: true}, nil
}

func (d *Dir) ReadDir(n int) ([]fs.DirEntry, error) {
	if n > 0 {
		return nil, io.EOF
	}
	return nil, nil
}

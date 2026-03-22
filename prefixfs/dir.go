package prefixfs

import (
	"io"
	"io/fs"
	"path"
	"time"

	"github.com/unstoppablemango/ihfs"
)

// dir is a synthetic read-only directory file for ancestor paths of the prefix.
// It implements [fs.ReadDirFile].
type dir struct {
	path  string // full path as opened (e.g. "a/b")
	child string // name of the next prefix component
	pos   int
}

func (d *dir) Stat() (ihfs.FileInfo, error) { return &dirInfo{name: path.Base(d.path)}, nil }
func (d *dir) Read([]byte) (int, error)     { return 0, d.pathError("read", ihfs.ErrInvalid) }
func (d *dir) Close() error                 { return nil }

func (d *dir) ReadDir(n int) ([]ihfs.DirEntry, error) {
	entries := []ihfs.DirEntry{&dirEntry{name: d.child}}
	if n <= 0 {
		if d.pos >= 1 {
			return nil, nil
		}
		d.pos = 1
		return entries, nil
	}

	if d.pos >= 1 {
		return nil, io.EOF
	}
	d.pos = 1
	return entries, nil
}

func (d *dir) pathError(op string, err error) error {
	return &ihfs.PathError{Op: op, Path: d.path, Err: err}
}

// dirEntry implements [fs.DirEntry] for synthetic ancestor directories.
type dirEntry struct {
	name string
}

func (e *dirEntry) Name() string                 { return e.name }
func (e *dirEntry) IsDir() bool                  { return true }
func (e *dirEntry) Type() ihfs.FileMode          { return fs.ModeDir }
func (e *dirEntry) Info() (ihfs.FileInfo, error) { return &dirInfo{name: e.name}, nil }

// dirInfo implements [fs.FileInfo] for synthetic ancestor directories.
type dirInfo struct {
	name string
}

func (i *dirInfo) Name() string        { return i.name }
func (i *dirInfo) Size() int64         { return 0 }
func (i *dirInfo) Mode() ihfs.FileMode { return fs.ModeDir | 0o555 }
func (i *dirInfo) ModTime() time.Time  { return time.Time{} }
func (i *dirInfo) IsDir() bool         { return true }
func (i *dirInfo) Sys() any            { return nil }

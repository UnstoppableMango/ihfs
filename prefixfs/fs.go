// Package prefixfs provides a filesystem implementation that makes an underlying
// filesystem accessible only under a given prefix path. It is the inverse of [fs.Sub].
package prefixfs

import (
	"io"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/unstoppablemango/ihfs"
)

// Fs wraps an FS and makes it accessible only under a prefix path.
// It is the inverse of [fs.Sub]: where Sub strips a prefix from paths,
// Fs adds one, so the wrapped FS is reachable only under the given prefix.
type Fs struct {
	fsys   ihfs.FS
	prefix string
}

// New creates a new Fs that makes fsys accessible under prefix.
// prefix must be a valid, non-root path (no leading slashes, no "..").
func New(fsys ihfs.FS, prefix string) (*Fs, error) {
	prefix = path.Clean(prefix)
	if !fs.ValidPath(prefix) || prefix == "." {
		return nil, &ihfs.PathError{Op: "prefix", Path: prefix, Err: ihfs.ErrInvalid}
	}
	return &Fs{fsys: fsys, prefix: prefix}, nil
}

// Base implements [ihfs.Decorator].
func (f *Fs) Base() ihfs.FS {
	return f.fsys
}

// Open implements [fs.FS].
func (f *Fs) Open(name string) (ihfs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &ihfs.PathError{Op: "open", Path: name, Err: ihfs.ErrInvalid}
	}

	if name == "." || strings.HasPrefix(f.prefix, name+"/") {
		return &dir{
			name:  path.Base(name),
			child: f.childComponent(name),
		}, nil
	}

	if name == f.prefix {
		return f.fsys.Open(".")
	}

	if strings.HasPrefix(name, f.prefix+"/") {
		return f.fsys.Open(name[len(f.prefix)+1:])
	}

	return nil, &ihfs.PathError{Op: "open", Path: name, Err: ihfs.ErrNotExist}
}

// Stat implements [ihfs.StatFS].
func (f *Fs) Stat(name string) (ihfs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, &ihfs.PathError{Op: "stat", Path: name, Err: ihfs.ErrInvalid}
	}

	if name == "." || strings.HasPrefix(f.prefix, name+"/") {
		return &dirInfo{name: path.Base(name)}, nil
	}

	if name == f.prefix {
		return fs.Stat(f.fsys, ".")
	}

	if strings.HasPrefix(name, f.prefix+"/") {
		return fs.Stat(f.fsys, name[len(f.prefix)+1:])
	}

	return nil, &ihfs.PathError{Op: "stat", Path: name, Err: ihfs.ErrNotExist}
}

// ReadDir implements [ihfs.ReadDirFS].
func (f *Fs) ReadDir(name string) ([]ihfs.DirEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &ihfs.PathError{Op: "readdir", Path: name, Err: ihfs.ErrInvalid}
	}

	if name == "." || strings.HasPrefix(f.prefix, name+"/") {
		return []ihfs.DirEntry{&dir{name: f.childComponent(name)}}, nil
	}

	if name == f.prefix {
		return fs.ReadDir(f.fsys, ".")
	}

	if strings.HasPrefix(name, f.prefix+"/") {
		return fs.ReadDir(f.fsys, name[len(f.prefix)+1:])
	}

	return nil, &ihfs.PathError{Op: "readdir", Path: name, Err: ihfs.ErrNotExist}
}

func (f *Fs) childComponent(name string) string {
	if name == "." {
		return firstComponent(f.prefix)
	}
	return firstComponent(f.prefix[len(name)+1:])
}

func firstComponent(p string) string {
	first, _, _ := strings.Cut(p, "/")
	return first
}

// dir is a synthetic read-only directory for ancestor paths of the prefix.
// It implements both [fs.ReadDirFile] and [fs.DirEntry].
type dir struct {
	name  string
	child string
	pos   int
}

func (d *dir) Stat() (ihfs.FileInfo, error) { return &dirInfo{name: d.name}, nil }
func (d *dir) Read([]byte) (int, error)     { return 0, d.error("read", ihfs.ErrInvalid) }
func (d *dir) Close() error                 { return nil }
func (d *dir) Name() string                 { return d.name }
func (d *dir) IsDir() bool                  { return true }
func (d *dir) Type() ihfs.FileMode          { return fs.ModeDir }
func (d *dir) Info() (ihfs.FileInfo, error) { return &dirInfo{name: d.name}, nil }

func (d *dir) ReadDir(n int) ([]ihfs.DirEntry, error) {
	entries := []ihfs.DirEntry{&dir{name: d.child}}
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

func (d *dir) error(op string, err error) error {
	return &ihfs.PathError{Op: op, Path: d.name, Err: err}
}

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

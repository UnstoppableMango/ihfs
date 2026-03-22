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
		return f.newVirtualDir(name), nil
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
		return &virtualDirInfo{name: path.Base(name)}, nil
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
		return []ihfs.DirEntry{f.nextEntry(name)}, nil
	}

	if name == f.prefix {
		return fs.ReadDir(f.fsys, ".")
	}

	if strings.HasPrefix(name, f.prefix+"/") {
		return fs.ReadDir(f.fsys, name[len(f.prefix)+1:])
	}

	return nil, &ihfs.PathError{Op: "readdir", Path: name, Err: ihfs.ErrNotExist}
}

func (f *Fs) newVirtualDir(name string) *virtualDir {
	return &virtualDir{
		name:  path.Base(name),
		child: f.childComponent(name),
	}
}

func (f *Fs) nextEntry(name string) *virtualDirEntry {
	return &virtualDirEntry{name: f.childComponent(name)}
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

// virtualDir is a synthetic read-only directory for ancestor paths of the prefix.
type virtualDir struct {
	name  string
	child string
	pos   int
}

func (d *virtualDir) Stat() (ihfs.FileInfo, error) {
	return &virtualDirInfo{name: d.name}, nil
}

func (d *virtualDir) Read([]byte) (int, error) {
	return 0, &ihfs.PathError{Op: "read", Path: d.name, Err: ihfs.ErrInvalid}
}

func (d *virtualDir) Close() error {
	return nil
}

func (d *virtualDir) ReadDir(n int) ([]ihfs.DirEntry, error) {
	entries := []ihfs.DirEntry{&virtualDirEntry{name: d.child}}
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

// virtualDirInfo implements [fs.FileInfo] for synthetic ancestor directories.
type virtualDirInfo struct {
	name string
}

func (i *virtualDirInfo) Name() string        { return i.name }
func (i *virtualDirInfo) Size() int64         { return 0 }
func (i *virtualDirInfo) Mode() ihfs.FileMode { return fs.ModeDir | 0o555 }
func (i *virtualDirInfo) ModTime() time.Time  { return time.Time{} }
func (i *virtualDirInfo) IsDir() bool         { return true }
func (i *virtualDirInfo) Sys() any            { return nil }

// virtualDirEntry implements [fs.DirEntry] for synthetic ancestor directories.
type virtualDirEntry struct {
	name string
}

func (e *virtualDirEntry) Name() string                    { return e.name }
func (e *virtualDirEntry) IsDir() bool                     { return true }
func (e *virtualDirEntry) Type() ihfs.FileMode             { return fs.ModeDir }
func (e *virtualDirEntry) Info() (ihfs.FileInfo, error)    { return &virtualDirInfo{name: e.name}, nil }

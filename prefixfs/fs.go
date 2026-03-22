// Package prefixfs provides a filesystem implementation that makes an underlying
// filesystem accessible only under a given prefix path. It is the inverse of [fs.Sub].
package prefixfs

import (
	"io/fs"
	"path"
	"strings"
)

// Fs wraps an FS and makes it accessible only under a prefix path.
// It is the inverse of [fs.Sub]: where Sub strips a prefix from paths,
// Fs adds one, so the wrapped FS is reachable only under the given prefix.
type Fs struct {
	fsys   fs.FS
	prefix string
}

// New creates a new Fs that makes fsys accessible under prefix.
// prefix is cleaned via [path.Clean] before use, and must resolve to a
// valid, non-root [fs.ValidPath]. New panics if the cleaned prefix is invalid.
func New(fsys fs.FS, prefix string) *Fs {
	prefix = path.Clean(prefix)
	if !fs.ValidPath(prefix) || prefix == "." {
		panic("prefixfs: invalid prefix: " + prefix)
	}
	return &Fs{fsys: fsys, prefix: prefix}
}

// Base returns the underlying filesystem.
func (f *Fs) Base() fs.FS {
	return f.fsys
}

// Open implements [fs.FS].
func (f *Fs) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	if name == "." || strings.HasPrefix(f.prefix, name+"/") {
		return &dir{
			path:  name,
			child: f.childComponent(name),
		}, nil
	}

	if name == f.prefix {
		return f.fsys.Open(".")
	}

	if strings.HasPrefix(name, f.prefix+"/") {
		return f.fsys.Open(name[len(f.prefix)+1:])
	}

	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

// Stat implements [fs.StatFS].
func (f *Fs) Stat(name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
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

	return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
}

// ReadDir implements [fs.ReadDirFS].
func (f *Fs) ReadDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}

	if name == "." || strings.HasPrefix(f.prefix, name+"/") {
		return []fs.DirEntry{&dirEntry{name: f.childComponent(name)}}, nil
	}

	if name == f.prefix {
		return fs.ReadDir(f.fsys, ".")
	}

	if strings.HasPrefix(name, f.prefix+"/") {
		return fs.ReadDir(f.fsys, name[len(f.prefix)+1:])
	}

	return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
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

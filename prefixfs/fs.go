// Package prefixfs provides a filesystem implementation that makes an underlying
// filesystem accessible only under a given prefix path. It is the inverse of [fs.Sub].
package prefixfs

import (
	"io/fs"
	"path"
	"strings"

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
		return []ihfs.DirEntry{&dirEntry{name: f.childComponent(name)}}, nil
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

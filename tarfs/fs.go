package tarfs

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"sync"

	"github.com/unstoppablemango/ihfs/osfs"
)

// Fs represents a read-only file system backed by a tar archive.
type Fs struct {
	name string
	mux  sync.RWMutex
	tar  io.ReadCloser

	fs map[string]*fileData
}

// Open opens a tar file as a read-only file system.
func Open(name string) (*Fs, error) {
	return OpenFS(osfs.Default, name)
}

// OpenFS opens a tar file from the given file system as a read-only file system.
func OpenFS(fs fs.FS, name string) (*Fs, error) {
	if f, err := fs.Open(name); err != nil {
		return nil, err
	} else {
		return FromReader(name, f), nil
	}
}

func FromReader(name string, r io.Reader) *Fs {
	fs := &Fs{name: name, fs: map[string]*fileData{}}
	if rc, ok := r.(io.ReadCloser); ok {
		fs.tar = rc
	} else {
		fs.tar = io.NopCloser(r)
	}

	return fs
}

func (t *Fs) Close() error {
	t.mux.Lock()
	defer t.mux.Unlock()

	if t.tar == nil {
		return nil
	}

	err := t.tar.Close()
	t.tar = nil
	return err
}

// Name returns the name of the tar file backing this file system.
func (t *Fs) Name() string {
	return t.name
}

// Open implements [fs.FS].
func (t *Fs) Open(name string) (fs.File, error) {
	t.mux.RLock()
	defer t.mux.RUnlock()

	if fd, ok := t.fs[name]; ok {
		return fd.file(), nil
	}

	tr := tar.NewReader(t.tar)

	for {
		fd, err := next(tr)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		t.fs[fd.hdr.Name] = fd
		if fd.hdr.Name == name {
			return fd.file(), nil
		}
	}

	if err := t.Close(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("%s: %w", name, fs.ErrNotExist)
}

func next(tr *tar.Reader) (*fileData, error) {
	hdr, err := tr.Next()
	if err != nil {
		return nil, err
	}

	if data, err := io.ReadAll(tr); err != nil {
		return nil, err
	} else {
		return &fileData{hdr, data}, nil
	}
}

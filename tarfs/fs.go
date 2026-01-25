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
	tr   *tar.Reader

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
	tfs := &Fs{name: name, fs: map[string]*fileData{}}
	if rc, ok := r.(io.ReadCloser); ok {
		tfs.tar = rc
	} else {
		tfs.tar = io.NopCloser(r)
	}

	tfs.tr = tar.NewReader(tfs.tar)
	return tfs
}

func (t *Fs) Close() error {
	t.mux.Lock()
	defer t.mux.Unlock()

	if t.tar == nil {
		return nil
	}

	err := t.tar.Close()
	t.tar = nil
	t.tr = nil
	return err
}

// Name returns the name of the tar file backing this file system.
func (t *Fs) Name() string {
	return t.name
}

// Open implements [fs.FS].
func (t *Fs) Open(name string) (fs.File, error) {
	t.mux.RLock()

	if fd, ok := t.fs[name]; ok {
		t.mux.RUnlock()
		return fd.file(), nil
	}

	if t.tr == nil {
		t.mux.RUnlock()
		return nil, fmt.Errorf("%s: %w", name, fs.ErrNotExist)
	}

	t.mux.RUnlock()
	t.mux.Lock()

	// Check again in case another goroutine loaded it
	if fd, ok := t.fs[name]; ok {
		t.mux.Unlock()
		return fd.file(), nil
	}

	if t.tr == nil {
		t.mux.Unlock()
		return nil, fmt.Errorf("%s: %w", name, fs.ErrNotExist)
	}

	// Lazy-load entries until we find the requested file
	for {
		fd, err := next(t.tr)
		if err == io.EOF {
			// Reached end of tar archive, close it
			closeErr := t.tar.Close()
			t.tar = nil
			t.tr = nil
			t.mux.Unlock()
			if closeErr != nil {
				return nil, closeErr
			}
			return nil, fmt.Errorf("%s: %w", name, fs.ErrNotExist)
		}
		if err != nil {
			t.mux.Unlock()
			return nil, err
		}

		t.fs[fd.hdr.Name] = fd
		if fd.hdr.Name == name {
			t.mux.Unlock()
			return fd.file(), nil
		}
	}
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

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
	name  string
	cache *cache
	mux   sync.Mutex
	tar   io.ReadCloser
	tr    *tar.Reader
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

// FromReader creates a new Fs from an io.Reader containing a tar archive.
func FromReader(name string, r io.Reader) *Fs {
	tfs := &Fs{name: name, cache: newCache()}
	if rc, ok := r.(io.ReadCloser); ok {
		tfs.tar = rc
	} else {
		tfs.tar = io.NopCloser(r)
	}
	tfs.tr = tar.NewReader(tfs.tar)
	return tfs
}

// Close closes the underlying tar archive.
func (t *Fs) Close() error {
	return t.tar.Close()
}

// Name returns the name of the tar file backing this file system.
func (t *Fs) Name() string {
	return t.name
}

// Open implements [fs.FS].
func (t *Fs) Open(name string) (fs.File, error) {
	// Check cache first with read lock
	if file := t.cache.get(name); file != nil {
		return file.file(), nil
	}

	// Not in cache, read from tar (only one goroutine at a time)
	t.mux.Lock()
	defer t.mux.Unlock()

	// Check cache again in case another goroutine loaded it
	if file := t.cache.get(name); file != nil {
		return file.file(), nil
	}

	// Lazy-load entries until we find the requested file
	for {
		fd, err := next(t.tr)
		if err != nil {
			return nil, t.error(name, fs.ErrNotExist, err)
		}

		t.cache.set(fd.hdr.Name, fd)
		if fd.hdr.Name == name {
			return fd.file(), nil
		}
	}
}

func (t *Fs) error(name string, err, cause error) error {
	return &TarError{
		Archive: t.name,
		Name:    name,
		Err:     err,
		Cause:   cause,
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

type TarError struct {
	Archive, Name string
	Err, Cause    error
}

func (e *TarError) Error() string {
	return fmt.Sprintf(
		"%s(%s): %v: %v",
		e.Archive, e.Name, e.Err, e.Cause,
	)
}

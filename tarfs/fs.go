package tarfs

import (
	"archive/tar"
	"fmt"
	"io"
	"sync"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/osfs"
)

// TarFile represents a read-only file system backed by a tar archive.
// It will lazily buffer the contents of the tar archive as files are accessed.
//
// Entries are accessed in order and cached as they are read, so random access may be inefficient.
type TarFile struct {
	name   string
	cache  *cache
	mux    sync.Mutex
	tar    io.ReadCloser
	tr     *tar.Reader
	closed bool
}

// Open opens a tar file as a read-only file system.
func Open(name string) (*TarFile, error) {
	return OpenFS(osfs.Default, name)
}

// OpenFS opens a tar file from fs as a read-only file system.
func OpenFS(fs ihfs.FS, name string) (*TarFile, error) {
	if f, err := fs.Open(name); err != nil {
		return nil, err
	} else {
		return FromReader(name, f), nil
	}
}

// FromReader creates a new TarFile from an [io.Reader] containing a tar archive.
//
// FromReader takes ownership of r, reading from it as needed. If r is an
// [io.ReadCloser] it will be closed when either [r.Read] returns an error
// or [Close] is called.
//
// If r is not an [io.ReadCloser], it will be wrapped in [io.NopCloser].
func FromReader(name string, r io.Reader) *TarFile {
	tfs := &TarFile{name: name, cache: newCache()}
	if rc, ok := r.(io.ReadCloser); ok {
		tfs.tar = rc
	} else {
		tfs.tar = io.NopCloser(r)
	}
	tfs.tr = tar.NewReader(tfs.tar)
	return tfs
}

// Close closes the underlying tar archive.
func (t *TarFile) Close() error {
	t.mux.Lock()
	defer t.mux.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	return t.tar.Close()
}

// Name returns the name of the tar file backing this file system.
func (t *TarFile) Name() string {
	return t.name
}

// Open implements [ihfs.FS].
func (t *TarFile) Open(name string) (ihfs.File, error) {
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

	if t.closed {
		return nil, t.notExist(name, ihfs.ErrClosed)
	}

	// Lazy-load entries until we find the requested file
	for {
		fd, err := next(t.tr)
		if err == io.EOF {
			if closeErr := t.close(); closeErr != nil {
				return nil, t.notExist(name, closeErr)
			}
			return nil, t.notExist(name, err)
		}
		if err != nil {
			return nil, t.notExist(name, err)
		}

		t.cache.set(fd.hdr.Name, fd)
		if fd.hdr.Name == name {
			return fd.file(), nil
		}
	}
}

func (t *TarFile) close() error {
	t.closed = true
	return t.tar.Close()
}

func (t *TarFile) notExist(name string, cause error) error {
	return t.error(name, ihfs.ErrNotExist, cause)
}

func (t *TarFile) error(name string, err, cause error) error {
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

func (e *TarError) Unwrap() []error {
	return []error{e.Err, e.Cause}
}

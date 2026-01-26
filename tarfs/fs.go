package tarfs

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"strings"
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
	cache  *cache // TODO: corfs with infinite timeout?
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
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	return FromReader(name, f), nil
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
	// Handle root directory before fs.ValidPath check (which rejects ".")
	if name == "." {
		// Return root directory - need to load all entries first
		t.mux.Lock()
		if !t.closed {
			for {
				fd, err := next(t.tr)
				if err == io.EOF {
					// Close the tar reader since we've fully read it
					if closeErr := t.close(); closeErr != nil {
						t.mux.Unlock()
						return nil, t.error(".", ihfs.ErrInvalid, closeErr)
					}
					break
				}
				if err != nil {
					t.mux.Unlock()
					return nil, t.error(".", ihfs.ErrInvalid, err)
				}
				name := fd.hdr.Name
				t.cache.set(name, fd)
				// Normalize directory names so callers using fs.ValidPath
				// can open "dir" and still get the real directory header
				if fd.hdr.Typeflag == tar.TypeDir && strings.HasSuffix(name, "/") {
					trimmed := strings.TrimSuffix(name, "/")
					if trimmed != "" {
						t.cache.set(trimmed, fd)
					}
				}
			}
		}
		t.mux.Unlock()

		// Return a synthetic directory for root
		return &File{
			hdr: &tar.Header{
				Name:     ".",
				Typeflag: tar.TypeDir,
				Mode:     0755,
			},
			name:  ".",
			cache: t.cache,
			r:     bytes.NewReader(nil),
		}, nil
	}

	if !fs.ValidPath(name) {
		return nil, t.invalid(name)
	}

	if file := t.cache.get(name); file != nil {
		return file.file(t.cache), nil
	}

	// Not in cache, read from tar (only one goroutine at a time)
	t.mux.Lock()
	defer t.mux.Unlock()

	// Check cache again in case another goroutine loaded it
	if file := t.cache.get(name); file != nil {
		return file.file(t.cache), nil
	}

	if t.closed {
		// Check if this is a synthetic directory
		prefix := name + "/"
		for _, fd := range t.cache.all() {
			if strings.HasPrefix(fd.hdr.Name, prefix) {
				// This is a valid directory - return synthetic entry
				return &File{
					hdr: &tar.Header{
						Name:     name,
						Typeflag: tar.TypeDir,
						Mode:     0755,
					},
					name:  name,
					cache: t.cache,
					r:     bytes.NewReader(nil),
				}, nil
			}
		}
		return nil, t.notExist(name, ihfs.ErrClosed)
	}

	// Lazy-load entries until we find the requested file
	for {
		hdr, err := t.tr.Next()
		if err == io.EOF {
			if closeErr := t.close(); closeErr != nil {
				return nil, t.notExist(name, closeErr)
			}
			// Check if this is a synthetic directory
			prefix := name + "/"
			for _, fd := range t.cache.all() {
				if strings.HasPrefix(fd.hdr.Name, prefix) {
					// This is a valid directory - return synthetic entry
					return &File{
						hdr: &tar.Header{
							Name:     name,
							Typeflag: tar.TypeDir,
							Mode:     0755,
						},
						name:  name,
						cache: t.cache,
						r:     bytes.NewReader(nil),
					}, nil
				}
			}
			return nil, t.notExist(name, err)
		}
		if err != nil {
			return nil, t.notExist(name, err)
		}

		entryName := fd.hdr.Name
		t.cache.set(entryName, fd)
		// Normalize directory names so callers using fs.ValidPath
		// can open "dir" and still get the real directory header
		if fd.hdr.Typeflag == tar.TypeDir && strings.HasSuffix(entryName, "/") {
			trimmed := strings.TrimSuffix(entryName, "/")
			if trimmed != "" {
				t.cache.set(trimmed, fd)
			}
		}
		if fd.hdr.Name == name || (fd.hdr.Typeflag == tar.TypeDir && strings.TrimSuffix(fd.hdr.Name, "/") == name) {
			return fd.file(t.cache), nil
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

func (t *TarFile) invalid(name string) error {
	return t.error(name, ihfs.ErrInvalid, nil)
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

	data, err := io.ReadAll(tr)
	if err != nil {
		return nil, err
	}
	return &fileData{hdr, data}, nil
}

// TarError represents an error that occurred while accessing a file in a tar archive.
type TarError struct {
	Archive, Name string
	Err, Cause    error
}

func (e *TarError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf(
			"%s(%s): %v: %v",
			e.Archive, e.Name, e.Err, e.Cause,
		)
	}
	return fmt.Sprintf("%s(%s): %v", e.Archive, e.Name, e.Err)
}

func (e *TarError) Unwrap() []error {
	return []error{e.Err, e.Cause}
}

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
)

type Fs struct {
	cache *cache
	mux   sync.Mutex
	tr    *tar.Reader
}

// FromReader creates a new TarFile from an [io.Reader] containing a tar archive.
//
// FromReader takes ownership of r, reading from it as needed. If r is an
// [io.ReadCloser] it will be closed when either [r.Read] returns an error
// or [Close] is called.
//
// If r is not an [io.ReadCloser], it will be wrapped in [io.NopCloser].
func FromReader(r io.Reader) *Fs {
	return &Fs{
		cache: newCache(),
		tr:    tar.NewReader(r),
	}
}

// Open implements [ihfs.FS].
func (t *Fs) Open(name string) (ihfs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  ihfs.ErrInvalid,
		}
	}

	// Non-directory files in the cache need no further work.
	if file := t.cache.get(name); file != nil && file.hdr.Typeflag != tar.TypeDir {
		return file.file(t.cache), nil
	}

	// Directory opens (cache hit or miss) always take the lock so drainIntoCache
	// can be called safely if the archive has not yet been fully read.
	t.mux.Lock()
	defer t.mux.Unlock()

	// Re-check cache under lock (handles both cache misses and directory hits).
	if file := t.cache.get(name); file != nil {
		if file.hdr.Typeflag == tar.TypeDir {
			if err := t.drainIntoCache(); err != nil {
				return nil, &fs.PathError{
					Op:   "open",
					Path: name,
					Err:  ihfs.ErrNotExist,
				}
			}
		}

		return file.file(t.cache), nil
	}

	// Lazy-load entries until we find the requested file
	for {
		fd, err := next(t.tr)
		if err == io.EOF {
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
		}
		if err != nil {
			return nil, &fs.PathError{
				Op:   "open",
				Path: name,
				Err:  ihfs.ErrNotExist,
			}
		}

		cacheKey := fd.hdr.Name
		if fd.hdr.Typeflag == tar.TypeDir {
			cacheKey = strings.TrimSuffix(cacheKey, "/")
		}
		if cacheKey != "" {
			t.cache.set(cacheKey, fd)
		}
		if cacheKey == name {
			if fd.hdr.Typeflag == tar.TypeDir {
				// Drain remaining entries so ReadDir returns a complete listing.
				if err := t.drainIntoCache(); err != nil {
					return nil, &fs.PathError{
						Op:   "open",
						Path: name,
						Err:  ihfs.ErrNotExist,
					}
				}
			}
			return fd.file(t.cache), nil
		}
	}
}

// drainIntoCache reads all remaining entries from the tar stream into the cache.
// The caller must hold t.mux and t.closed must be false.
// drainIntoCache calls t.close() after reaching EOF.
func (t *Fs) drainIntoCache() error {
	for {
		fd, err := next(t.tr)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		cacheKey := fd.hdr.Name
		if fd.hdr.Typeflag == tar.TypeDir {
			cacheKey = strings.TrimSuffix(cacheKey, "/")
		}
		if cacheKey != "" {
			t.cache.set(cacheKey, fd)
		}
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

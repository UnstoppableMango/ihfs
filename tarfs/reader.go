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

type Reader struct {
	cache *cache
	mux   sync.Mutex
	tr    *tar.Reader
}

// FromReader creates a new TarFile from an [io.Reader] containing a tar archive.
func FromReader(r io.Reader) *Reader {
	return &Reader{
		cache: newCache(),
		tr:    tar.NewReader(r),
	}
}

// Open implements [ihfs.FS].
func (t *Reader) Open(name string) (ihfs.File, error) {
	if name == "." {
		t.mux.Lock()
		defer t.mux.Unlock()
		if err := t.drainIntoCache(); err != nil {
			return nil, &fs.PathError{Op: "open", Path: ".", Err: err}
		}
		return &File{
			hdr:   &tar.Header{Name: ".", Typeflag: tar.TypeDir, Mode: 0755},
			name:  ".",
			cache: t.cache,
			r:     bytes.NewReader(nil),
		}, nil
	}

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
				Err:  fmt.Errorf("%w: %w", ihfs.ErrNotExist, err),
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
// The caller must hold t.mux.
func (t *Reader) drainIntoCache() error {
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

package tarfs

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"strings"
	"sync"

	"github.com/unstoppablemango/ihfs"
)

type Fs struct {
	cache *sync.Map
	mux   sync.Mutex
	tr    *tar.Reader
}

// FromReader creates a new TarFile from an [io.Reader] containing a tar archive.
func FromReader(r io.Reader) *Fs {
	return &Fs{
		cache: &sync.Map{},
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
	if f, ok := t.cache.Load(name); ok {
		return f.(fs.File), nil
	}

	return t.until(name)
}

func (t *Fs) ReadDir(name string) ([]fs.DirEntry, error) {
	if _, err := t.until(""); err != nil {
		return nil, err
	}
	return dirEntries(t.cache, name), nil
}

func (t *Fs) until(name string) (fs.File, error) {
	t.mux.Lock()
	defer t.mux.Unlock()

	for {
		hdr, err := t.tr.Next()
		if err == io.EOF {
			return nil, fs.ErrNotExist
		}
		if err != nil {
			return nil, err
		}

		f, err := t.read(name, hdr)
		if err != nil {
			return nil, err
		}

		key := hdr.Name
		t.cache.Store(key, f)
		if name == strings.TrimLeft(key, "/") {
			return f, nil
		}
	}
}

func (t *Fs) read(name string, hdr *tar.Header) (entry, error) {
	if hdr.Typeflag == tar.TypeDir {
		return &Dir{name: name, fsys: t}, nil
	}

	data, err := io.ReadAll(t.tr)
	if err != nil {
		return nil, err
	}

	return &File{
		name: name,
		hdr:  hdr,
		r:    bytes.NewReader(data),
	}, nil
}

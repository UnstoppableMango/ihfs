package tarfs

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"sync"

	"github.com/unstoppablemango/ihfs/osfs"
)

type Fs struct {
	name string
	mux  sync.Mutex
	tr   *tar.Reader
	fs   map[string]fileData
}

func Open(name string) (*Fs, error) {
	return OpenFS(osfs.Default, name)
}

func OpenFS(fs fs.FS, name string) (*Fs, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}

	return &Fs{
		name: name,
		tr:   tar.NewReader(f),
		fs:   map[string]fileData{},
	}, nil
}

func (t *Fs) Name() string {
	return t.name
}

// Open implements [fs.FS].
func (t *Fs) Open(name string) (fs.File, error) {
	t.mux.Lock()
	defer t.mux.Unlock()

	if fd, ok := t.fs[name]; ok {
		return fd.file(), nil
	}

	for {
		fd, err := t.next()
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

	return nil, fmt.Errorf("%s: %w", name, fs.ErrNotExist)
}

// next reads the next entry from the tar reader.
// Caller must hold t.mux.
func (t *Fs) next() (fileData, error) {
	hdr, err := t.tr.Next()
	if err != nil {
		return fileData{}, err
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(t.tr); err != nil {
		return fileData{}, err
	}

	return fileData{
		hdr:  hdr,
		data: buf.Bytes(),
	}, nil
}

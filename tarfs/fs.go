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
	fs   map[string]*File
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
		fs:   map[string]*File{},
	}, nil
}

func (t *Fs) Name() string {
	return t.name
}

// Open implements [TarFile].
func (t *Fs) Open(name string) (fs.File, error) {
	if f, ok := t.fs[name]; ok {
		return f, nil
	}

	for {
		file, err := t.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		t.fs[file.Name()] = file
		if file.Name() == name {
			return file, nil
		}
	}

	return nil, fmt.Errorf("%s: %w", name, fs.ErrNotExist)
}

func (t *Fs) next() (*File, error) {
	t.mux.Lock()
	defer t.mux.Unlock()

	hdr, err := t.tr.Next()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(t.tr); err != nil {
		return nil, err
	}

	return &File{hdr, &buf}, nil
}

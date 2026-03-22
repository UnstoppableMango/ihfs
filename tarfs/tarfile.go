package tarfs

import (
	"errors"
	"io/fs"
	"sync/atomic"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/osfs"
)

// TarFile represents a read-only file system backed by a tar archive.
// It will lazily buffer the contents of the tar archive as files are accessed.
// Opening a directory entry causes all remaining archive entries to be read
// eagerly so that ReadDir returns a complete listing.
//
// Memory: All file content read from the archive is held in memory to support
// random access on a sequential stream. For large archives this can be
// significant. Closing TarFile does not release cached content because open
// File handles may still hold references to the cache.
//
// Entries are accessed in order and cached as they are read, so random access may be inefficient.
type TarFile struct {
	*Fs

	name   string
	file   fs.File
	closed atomic.Bool
}

// Open opens a tar file as a read-only file system.
func Open(name string) (*TarFile, error) {
	return OpenFS(osfs.Default, name)
}

// OpenFS opens a tar file from fs as a read-only file system.
func OpenFS(fsys ihfs.FS, name string) (*TarFile, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}
	return &TarFile{
		Fs:   FromReader(f),
		name: name,
		file: f,
	}, nil
}

// Name returns the name of the backing tar file.
func (t *TarFile) Name() string {
	return t.name
}

// Open implements [ihfs.FS], adding closed-state checking and [TarError] wrapping.
func (t *TarFile) Open(name string) (ihfs.File, error) {
	if t.closed.Load() {
		return nil, &TarError{
			Archive: t.name,
			Name:    name,
			Err:     fs.ErrNotExist,
		}
	}
	file, err := t.Fs.Open(name)
	if err != nil {
		return nil, t.wrapErr(name, err)
	}
	return file, nil
}

// Close closes the underlying tar archive.
func (t *TarFile) Close() error {
	if t.closed.Swap(true) {
		return nil
	}
	return t.file.Close()
}

func (t *TarFile) wrapErr(name string, err error) error {
	if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
		err = pathErr.Err
	}
	return &TarError{Archive: t.name, Name: name, Err: err}
}

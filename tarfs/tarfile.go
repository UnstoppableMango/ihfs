package tarfs

import (
	"errors"
	"io"
	"io/fs"
	"sync/atomic"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/osfs"
)

// TarFile represents a file system backed by a tar archive.
// It supports either reading or writing, but never both.
// For reading, use [Open] or [OpenFS]. For writing, use [Create] or [CreateFS].
// Read-only TarFiles are backed by [Reader].
// Write-only TarFiles are backed by [Writer].
type TarFile struct {
	name   string
	file   fs.File
	closed atomic.Bool
	fs     ihfs.CloserFS
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
		fs:   ihfs.NopCloser(FromReader(f)),
		name: name,
		file: f,
	}, nil
}

func Create(name string) (*TarFile, error) {
	return CreateFS(osfs.Default, name)
}

func CreateFS(fsys ihfs.FS, name string) (*TarFile, error) {
	f, err := ihfs.Create(fsys, name)
	if err != nil {
		return nil, err
	}

	w, ok := f.(io.Writer)
	if !ok {
		_ = f.Close()
		return nil, errors.New("file does not support writing")
	}

	return &TarFile{
		fs:   NewWriter(w),
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
		return nil, t.wrapErr(name, fs.ErrNotExist)
	}
	file, err := t.fs.Open(name)
	if err != nil {
		return nil, t.wrapErr(name, err)
	}
	return file, nil
}

func (t *TarFile) OpenFile(name string, flag int, perm fs.FileMode) (ihfs.File, error) {
	if t.closed.Load() {
		return nil, t.wrapErr(name, fs.ErrNotExist)
	}
	return ihfs.OpenFile(t.fs, name, flag, perm)
}

// Close closes the underlying tar archive.
func (t *TarFile) Close() error {
	if t.closed.Swap(true) {
		return nil
	}
	return errors.Join(
		t.fs.Close(),
		t.file.Close(),
	)
}

func (t *TarFile) Create(name string) (ihfs.File, error) {
	if t.closed.Load() {
		return nil, t.wrapErr(name, fs.ErrNotExist)
	}
	return ihfs.Create(t.fs, name)
}

func (t *TarFile) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if t.closed.Load() {
		return t.wrapErr(name, fs.ErrNotExist)
	}
	return ihfs.WriteFile(t.fs, name, data, perm)
}

func (t *TarFile) Mkdir(name string, perm fs.FileMode) error {
	if t.closed.Load() {
		return t.wrapErr(name, fs.ErrNotExist)
	}
	return ihfs.Mkdir(t.fs, name, perm)
}

func (t *TarFile) MkdirAll(name string, perm fs.FileMode) error {
	if t.closed.Load() {
		return t.wrapErr(name, fs.ErrNotExist)
	}
	return ihfs.MkdirAll(t.fs, name, perm)
}

func (t *TarFile) Copy(dir string, fsys ihfs.FS) error {
	if t.closed.Load() {
		return t.wrapErr(dir, fs.ErrNotExist)
	}
	return ihfs.Copy(t.fs, dir, fsys)
}

func (t *TarFile) Symlink(oldname, newname string) error {
	if t.closed.Load() {
		return t.wrapErr(newname, fs.ErrNotExist)
	}
	return ihfs.Symlink(t.fs, oldname, newname)
}

func (t *TarFile) wrapErr(name string, err error) error {
	if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
		err = pathErr.Err
	}
	return &TarError{
		Archive: t.name,
		Name:    name,
		Err:     err,
	}
}

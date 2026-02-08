package testfs

import (
	"fmt"

	"github.com/unstoppablemango/ihfs"
)

// File is a configurable test file implementation.
type File struct {
	name string

	CloseFunc        func() error
	ReadFunc         func(p []byte) (n int, err error)
	StatFunc         func() (ihfs.FileInfo, error)
	SeekFunc         func(offset int64, whence int) (int64, error)
	WriteFunc        func(p []byte) (n int, err error)
	ReadDirFunc      func(n int) ([]ihfs.DirEntry, error)
	ReadAtFunc       func(p []byte, off int64) (int, error)
	WriteAtFunc      func(p []byte, off int64) (int, error)
	WriteStringFunc  func(s string) (int, error)
	SyncFunc         func() error
	TruncateFunc     func(size int64) error
	ReadDirNamesFunc func(n int) ([]string, error)
}

// Close implements [ihfs.File].
func (f *File) Close() error {
	if f.CloseFunc != nil {
		return f.CloseFunc()
	}
	return nil
}

// Name returns the file's name.
func (f *File) Name() string {
	return f.name
}

// Read implements [io.Reader].
func (f *File) Read(p []byte) (n int, err error) {
	if f.ReadFunc != nil {
		return f.ReadFunc(p)
	}
	return 0, fmt.Errorf("read: %w", ErrNoMocks)
}

// Stat implements [fs.File].
func (f *File) Stat() (ihfs.FileInfo, error) {
	if f.StatFunc != nil {
		return f.StatFunc()
	}
	return nil, fmt.Errorf("stat: %w", ErrNoMocks)
}

// Seek implements [io.Seeker].
func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.SeekFunc != nil {
		return f.SeekFunc(offset, whence)
	}
	return 0, fmt.Errorf("seek: %w", ErrNoMocks)
}

// Write implements [io.Writer].
func (f *File) Write(p []byte) (n int, err error) {
	if f.WriteFunc != nil {
		return f.WriteFunc(p)
	}
	return 0, fmt.Errorf("write: %w", ErrNoMocks)
}

// ReadDir implements [fs.ReadDirFile].
func (f *File) ReadDir(n int) ([]ihfs.DirEntry, error) {
	if f.ReadDirFunc != nil {
		return f.ReadDirFunc(n)
	}
	return nil, fmt.Errorf("readdir: %w", ErrNoMocks)
}

// ReadAt implements [io.ReaderAt].
func (f *File) ReadAt(p []byte, off int64) (int, error) {
	if f.ReadAtFunc != nil {
		return f.ReadAtFunc(p, off)
	}
	return 0, fmt.Errorf("readat: %w", ErrNoMocks)
}

// WriteAt implements [io.WriterAt].
func (f *File) WriteAt(p []byte, off int64) (int, error) {
	if f.WriteAtFunc != nil {
		return f.WriteAtFunc(p, off)
	}
	return 0, fmt.Errorf("writeat: %w", ErrNoMocks)
}

// WriteString implements [io.StringWriter].
func (f *File) WriteString(s string) (int, error) {
	if f.WriteStringFunc != nil {
		return f.WriteStringFunc(s)
	}
	return 0, fmt.Errorf("writestring: %w", ErrNoMocks)
}

// Sync implements [ihfs.SyncFile].
func (f *File) Sync() error {
	if f.SyncFunc != nil {
		return f.SyncFunc()
	}
	return fmt.Errorf("sync: %w", ErrNoMocks)
}

// Truncate implements [ihfs.TruncateFile].
func (f *File) Truncate(size int64) error {
	if f.TruncateFunc != nil {
		return f.TruncateFunc(size)
	}
	return fmt.Errorf("truncate: %w", ErrNoMocks)
}

// ReadDirNames implements [ihfs.ReadDirNamesFile].
func (f *File) ReadDirNames(n int) ([]string, error) {
	if f.ReadDirNamesFunc != nil {
		return f.ReadDirNamesFunc(n)
	}
	return nil, fmt.Errorf("readdirnames: %w", ErrNoMocks)
}

// DirEntry is a configurable test directory entry implementation.
type DirEntry struct {
	name string

	IsDirFunc func() bool
	TypeFunc  func() ihfs.FileMode
	InfoFunc  func() (ihfs.FileInfo, error)
}

// NewDirEntry creates a new [DirEntry] with the given name and directory flag.
func NewDirEntry(name string, isDir bool) *DirEntry {
	return &DirEntry{
		name:      name,
		IsDirFunc: func() bool { return isDir },
		TypeFunc:  func() ihfs.FileMode { return 0 },
		InfoFunc: func() (ihfs.FileInfo, error) {
			fi := NewFileInfo(name)
			fi.IsDirFunc = func() bool { return isDir }
			return fi, nil
		},
	}
}

// Name implements [fs.DirEntry].
func (d *DirEntry) Name() string {
	return d.name
}

// IsDir implements [fs.DirEntry].
func (d *DirEntry) IsDir() bool {
	return d.IsDirFunc()
}

// Type implements [fs.DirEntry].
func (d *DirEntry) Type() ihfs.FileMode {
	return d.TypeFunc()
}

// Info implements [fs.DirEntry].
func (d *DirEntry) Info() (ihfs.FileInfo, error) {
	if d.InfoFunc != nil {
		return d.InfoFunc()
	}
	return nil, fmt.Errorf("info: %w", ErrNoMocks)
}

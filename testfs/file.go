package testfs

import (
	"fmt"

	"github.com/unstoppablemango/ihfs"
)

type File struct {
	name string

	CloseFunc   func() error
	ReadFunc    func(p []byte) (n int, err error)
	StatFunc    func() (ihfs.FileInfo, error)
	SeekFunc    func(offset int64, whence int) (int64, error)
	WriteFunc   func(p []byte) (n int, err error)
	ReadDirFunc func(n int) ([]ihfs.DirEntry, error)
}

func (f *File) Close() error {
	return f.CloseFunc()
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Read(p []byte) (n int, err error) {
	if f.ReadFunc != nil {
		return f.ReadFunc(p)
	}
	return 0, fmt.Errorf("read: %w", ErrNotImplemented)
}

func (f *File) Stat() (ihfs.FileInfo, error) {
	if f.StatFunc != nil {
		return f.StatFunc()
	}
	return nil, fmt.Errorf("stat: %w", ErrNotImplemented)
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.SeekFunc != nil {
		return f.SeekFunc(offset, whence)
	}
	return 0, fmt.Errorf("seek: %w", ErrNotImplemented)
}

func (f *File) Write(p []byte) (n int, err error) {
	if f.WriteFunc != nil {
		return f.WriteFunc(p)
	}
	return 0, fmt.Errorf("write: %w", ErrNotImplemented)
}

func (f *File) ReadDir(n int) ([]ihfs.DirEntry, error) {
	if f.ReadDirFunc != nil {
		return f.ReadDirFunc(n)
	}
	return nil, fmt.Errorf("readdir: %w", ErrNotImplemented)
}

type DirEntry struct {
	name string

	IsDirFunc func() bool
	TypeFunc  func() ihfs.FileMode
	InfoFunc  func() (ihfs.FileInfo, error)
}

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

func (d *DirEntry) Name() string {
	return d.name
}

func (d *DirEntry) IsDir() bool {
	return d.IsDirFunc()
}

func (d *DirEntry) Type() ihfs.FileMode {
	return d.TypeFunc()
}

func (d *DirEntry) Info() (ihfs.FileInfo, error) {
	if d.InfoFunc != nil {
		return d.InfoFunc()
	}
	return nil, fmt.Errorf("info: %w", ErrNotImplemented)
}

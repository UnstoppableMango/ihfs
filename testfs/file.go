package testfs

import "github.com/unstoppablemango/ihfs"

type File struct {
	CloseFunc   func() error
	NameFunc    func() string
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
	return f.NameFunc()
}

func (f *File) Read(p []byte) (n int, err error) {
	return f.ReadFunc(p)
}

func (f *File) Stat() (ihfs.FileInfo, error) {
	return f.StatFunc()
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	return f.SeekFunc(offset, whence)
}

func (f *File) Write(p []byte) (n int, err error) {
	return f.WriteFunc(p)
}

func (f *File) ReadDir(n int) ([]ihfs.DirEntry, error) {
	return f.ReadDirFunc(n)
}

type BoringFile struct {
	CloseFunc func() error
	ReadFunc  func(p []byte) (n int, err error)
	StatFunc  func() (ihfs.FileInfo, error)
}

func (f *BoringFile) Close() error {
	return f.CloseFunc()
}

func (f *BoringFile) Read(p []byte) (n int, err error) {
	return f.ReadFunc(p)
}

func (f *BoringFile) Stat() (ihfs.FileInfo, error) {
	return f.StatFunc()
}

type DirEntry struct {
	NameFunc  func() string
	IsDirFunc func() bool
	TypeFunc  func() ihfs.FileMode
	InfoFunc  func() (ihfs.FileInfo, error)
}

func NewDirEntry(name string, isDir bool) *DirEntry {
	return &DirEntry{
		NameFunc:  func() string { return name },
		IsDirFunc: func() bool { return isDir },
		TypeFunc:  func() ihfs.FileMode { return 0 },
		InfoFunc: func() (ihfs.FileInfo, error) {
			fi := NewFileInfo()
			fi.NameFunc = func() string { return name }
			fi.IsDirFunc = func() bool { return isDir }
			return fi, nil
		},
	}
}

func (d *DirEntry) Name() string {
	return d.NameFunc()
}

func (d *DirEntry) IsDir() bool {
	return d.IsDirFunc()
}

func (d *DirEntry) Type() ihfs.FileMode {
	return d.TypeFunc()
}

func (d *DirEntry) Info() (ihfs.FileInfo, error) {
	return d.InfoFunc()
}

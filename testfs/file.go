package testfs

import "github.com/unstoppablemango/ihfs"

type File struct {
	CloseFunc func() error
	NameFunc  func() string
	ReadFunc  func(p []byte) (n int, err error)
	StatFunc  func() (ihfs.FileInfo, error)
	SeekFunc  func(offset int64, whence int) (int64, error)
	WriteFunc func(p []byte) (n int, err error)
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

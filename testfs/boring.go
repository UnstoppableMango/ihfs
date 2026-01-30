package testfs

import (
	"fmt"

	"github.com/unstoppablemango/ihfs"
)

type BoringFs struct {
	OpenFunc func(string) (ihfs.File, error)
}

func (fs BoringFs) Open(name string) (ihfs.File, error) {
	if fs.OpenFunc != nil {
		return fs.OpenFunc(name)
	}
	return nil, fmt.Errorf("open: %w", ErrNotImplemented)
}

type BoringFile struct {
	CloseFunc func() error
	ReadFunc  func(p []byte) (n int, err error)
	StatFunc  func() (ihfs.FileInfo, error)
}

func (f *BoringFile) Close() error {
	if f.CloseFunc != nil {
		return f.CloseFunc()
	}
	return fmt.Errorf("close: %w", ErrNotImplemented)
}

func (f *BoringFile) Read(p []byte) (n int, err error) {
	if f.ReadFunc != nil {
		return f.ReadFunc(p)
	}
	return 0, fmt.Errorf("read: %w", ErrNotImplemented)
}

func (f *BoringFile) Stat() (ihfs.FileInfo, error) {
	if f.StatFunc != nil {
		return f.StatFunc()
	}
	return nil, fmt.Errorf("stat: %w", ErrNotImplemented)
}

package testfs

import (
	"io/fs"
	"testing/fstest"

	"github.com/unmango/go/fopt"
	"github.com/unstoppablemango/ihfs"
)

type (
	Map = fstest.MapFS
)

type Fs struct {
	OpenFunc func(string) (ihfs.File, error)
	StatFunc func(string) (ihfs.FileInfo, error)
}

func New(opts ...Option) Fs {
	fs := Fs{
		OpenFunc: defaultOpenFunc,
		StatFunc: defaultStatFunc,
	}

	fopt.ApplyAll(&fs, opts)
	return fs
}

func (fs Fs) Open(name string) (ihfs.File, error) {
	return fs.OpenFunc(name)
}

func defaultOpenFunc(name string) (ihfs.File, error) {
	return nil, fs.ErrNotExist
}

func (fs Fs) Stat(name string) (ihfs.FileInfo, error) {
	return fs.StatFunc(name)
}

func defaultStatFunc(name string) (ihfs.FileInfo, error) {
	return nil, fs.ErrNotExist
}

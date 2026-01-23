package testfs

import (
	"io/fs"

	"github.com/unmango/go/fopt"
	"github.com/unstoppablemango/ihfs"
)

type Fs struct {
	OpenFunc func(string) (ihfs.File, error)
}

type Option func(*Fs)

func New(opts ...Option) Fs {
	fs := Fs{OpenFunc: defaultOpenFunc}
	fopt.ApplyAll(&fs, opts)
	return fs
}

func (fs Fs) Open(name string) (ihfs.File, error) {
	return fs.OpenFunc(name)
}

func defaultOpenFunc(name string) (ihfs.File, error) {
	return nil, fs.ErrNotExist
}

func WithOpen(fn func(string) (ihfs.File, error)) Option {
	return func(fs *Fs) {
		fs.OpenFunc = fn
	}
}

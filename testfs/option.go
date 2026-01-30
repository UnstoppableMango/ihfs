package testfs

import (
	"time"

	"github.com/unstoppablemango/ihfs"
)

type Option func(*Fs)

func WithOpen(fn func(string) (ihfs.File, error)) Option {
	return func(fs *Fs) {
		fs.OpenFunc = fn
	}
}

func WithStat(fn func(string) (ihfs.FileInfo, error)) Option {
	return func(fs *Fs) {
		fs.StatFunc = fn
	}
}

func WithCreate(fn func(string) (ihfs.File, error)) Option {
	return func(fs *Fs) {
		fs.CreateFunc = fn
	}
}

func WithWriteFile(fn func(string, []byte, ihfs.FileMode) error) Option {
	return func(fs *Fs) {
		fs.WriteFileFunc = fn
	}
}

func WithChmod(fn func(string, ihfs.FileMode) error) Option {
	return func(fs *Fs) {
		fs.ChmodFunc = fn
	}
}

func WithChown(fn func(string, int, int) error) Option {
	return func(fs *Fs) {
		fs.ChownFunc = fn
	}
}

func WithChtimes(fn func(string, time.Time, time.Time) error) Option {
	return func(fs *Fs) {
		fs.ChtimesFunc = fn
	}
}

func WithCopy(fn func(string, ihfs.FS) error) Option {
	return func(fs *Fs) {
		fs.CopyFunc = fn
	}
}

func WithMkdir(fn func(string, ihfs.FileMode) error) Option {
	return func(fs *Fs) {
		fs.MkdirFunc = fn
	}
}

func WithMkdirAll(fn func(string, ihfs.FileMode) error) Option {
	return func(fs *Fs) {
		fs.MkdirAllFunc = fn
	}
}

func WithMkdirTemp(fn func(string, string) (string, error)) Option {
	return func(fs *Fs) {
		fs.MkdirTempFunc = fn
	}
}

func WithRemove(fn func(string) error) Option {
	return func(fs *Fs) {
		fs.RemoveFunc = fn
	}
}

func WithRemoveAll(fn func(string) error) Option {
	return func(fs *Fs) {
		fs.RemoveAllFunc = fn
	}
}

func WithReadDir(fn func(string) ([]ihfs.DirEntry, error)) Option {
	return func(fs *Fs) {
		fs.ReadDirFunc = fn
	}
}

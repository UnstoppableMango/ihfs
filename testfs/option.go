package testfs

import "github.com/unstoppablemango/ihfs"

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

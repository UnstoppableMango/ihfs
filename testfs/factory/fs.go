package factory

import (
	"errors"

	"github.com/unstoppablemango/ihfs"
)

type (
	OpenFunc func(string) (ihfs.File, error)
	StatFunc func(string) (ihfs.FileInfo, error)
)

var ErrNotMocked = errors.New("operation has no mock")

type Fs struct {
	name string
	open []OpenFunc
	stat []StatFunc
}

func NewFs() *Fs {
	return &Fs{name: "testfs/factory"}
}

func (f *Fs) Named(name string) *Fs {
	f.name = name
	return f
}

func (f *Fs) Name() string {
	return f.name
}

func (f *Fs) WithOpen(open ...OpenFunc) *Fs {
	f.open = append(f.open, open...)
	return f
}

func (f *Fs) SetOpen(open ...OpenFunc) *Fs {
	f.open = open
	return f
}

func (f *Fs) Open(path string) (ihfs.File, error) {
	if len(f.open) == 0 {
		return nil, ErrNotMocked
	}

	open := f.open[0]
	f.open = f.open[1:]
	return open(path)
}

func (f *Fs) WithStat(stat ...StatFunc) *Fs {
	f.stat = append(f.stat, stat...)
	return f
}

func (f *Fs) SetStat(stat ...StatFunc) *Fs {
	f.stat = stat
	return f
}

func (f *Fs) Stat(path string) (ihfs.FileInfo, error) {
	if len(f.stat) == 0 {
		return nil, ErrNotMocked
	}

	stat := f.stat[0]
	f.stat = f.stat[1:]
	return stat(path)
}

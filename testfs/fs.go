package testfs

import (
	"github.com/unstoppablemango/ihfs"
)

type Fs struct {
	OpenFunc func(string) (ihfs.File, error)
}

func (fs Fs) Open(name string) (ihfs.File, error) {
	return fs.OpenFunc(name)
}

package ghfs

import (
	"github.com/google/go-github/v82/github"
	"github.com/unstoppablemango/ihfs"
)

type Fs struct {
	client *github.Client
}

func (*Fs) Name() string {
	return "github"
}

func (f *Fs) Open(name string) (ihfs.File, error) {
	return nil, nil
}

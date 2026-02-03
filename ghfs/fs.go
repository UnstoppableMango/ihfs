package ghfs

import (
	"context"

	"github.com/google/go-github/v82/github"
	"github.com/unstoppablemango/ihfs"
)

type (
	ContextFunc func(*Fs, ihfs.Operation) context.Context
)

type Fs struct {
	client *github.Client
	ctxFn  ContextFunc
}

func New() *Fs {
	return &Fs{
		client: github.NewClient(nil),
		ctxFn:  background,
	}
}

func (*Fs) Name() string {
	return "github"
}

func (f *Fs) Open(name string) (ihfs.File, error) {
	return nil, nil
}

func (f *Fs) context(op ihfs.Operation) context.Context {
	return f.ctxFn(f, op)
}

func background(*Fs, ihfs.Operation) context.Context {
	return context.Background()
}

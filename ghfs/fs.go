package ghfs

import (
	"context"
	"strings"

	"github.com/google/go-github/v82/github"
	"github.com/unmango/go/fopt"
	"github.com/unstoppablemango/ihfs"
)

type (
	ContextFunc func(*Fs, ihfs.Operation) context.Context
)

type Fs struct {
	client *github.Client
	ctxFn  ContextFunc
}

func New(options ...Option) *Fs {
	f := &Fs{
		client: github.NewClient(nil),
		ctxFn:  background,
	}

	fopt.ApplyAll(f, options)
	return f
}

func (*Fs) Name() string {
	return "github"
}

func (f *Fs) Open(name string) (ihfs.File, error) {
	parts := strings.Split(clean(name), "/")

	switch len(parts) {
	case 1:
		return &Owner{
			name: parts[0],
		}, nil
	case 2:
		return &Repository{
			owner: parts[0],
			name:  parts[1],
		}, nil
	case 4:
		return &Branch{
			owner:      parts[0],
			repository: parts[1],
			name:       parts[3],
		}, nil
	case 5:
		if parts[2] == "blob" {
			return &Content{
				owner:      parts[0],
				repository: parts[1],
				branch:     parts[3],
				name:       parts[4],
			}, nil
		}

		return &Release{
			owner:      parts[0],
			repository: parts[1],
			name:       parts[4],
		}, nil
	case 6:
		return &Asset{
			owner:      parts[0],
			repository: parts[1],
			release:    parts[4],
			name:       parts[5],
		}, nil
	}

	return nil, ihfs.ErrNotExist
}

func (f *Fs) setAuthToken(token string) {
	f.client = f.client.WithAuthToken(token)
}

func (f *Fs) context(op ihfs.Operation) context.Context {
	return f.ctxFn(f, op)
}

func background(*Fs, ihfs.Operation) context.Context {
	return context.Background()
}

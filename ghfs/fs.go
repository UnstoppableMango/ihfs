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
		return f.openOwner(parts[0])
	case 2:
		return f.openRepository(parts[0], parts[1])
	case 4:
		return f.openBranch(parts[0], parts[1], parts[3])
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
	}

	if len(parts) >= 6 {
		if parts[2] == "releases" {
			return &Asset{
				owner:      parts[0],
				repository: parts[1],
				release:    parts[4],
				name:       parts[5],
			}, nil
		}

		return &Content{
			owner:      parts[0],
			repository: parts[1],
			branch:     parts[3],
			name:       strings.Join(parts[4:], "/"),
		}, nil
	}

	return nil, &ihfs.PathError{
		Op:   "open",
		Path: name,
		Err:  ihfs.ErrNotExist,
	}
}

func (f *Fs) setAuthToken(token string) {
	f.client = f.client.WithAuthToken(token)
}

func (f *Fs) context(op ihfs.Operation) context.Context {
	return f.ctxFn(f, op)
}

func (f *Fs) openOwner(name string) (*Owner, error) {
	return &Owner{
		name: name,
	}, nil
}

func (f *Fs) openRepository(owner, name string) (*Repository, error) {
	return &Repository{
		owner: owner,
		name:  name,
	}, nil
}

func (f *Fs) openBranch(owner, repository, name string) (*Branch, error) {
	return &Branch{
		owner:      owner,
		repository: repository,
		name:       name,
	}, nil
}

func background(*Fs, ihfs.Operation) context.Context {
	return context.Background()
}

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
			return f.openContent(parts[0], parts[1], parts[3], parts[4])
		}
		return f.openRelease(parts[0], parts[1], parts[4])
	}

	if len(parts) >= 6 {
		if parts[2] == "releases" {
			return f.openAsset(parts[0], parts[1], parts[4], parts[5])
		}
		return f.openContent(parts[0], parts[1], parts[3],
			strings.Join(parts[4:], "/"),
		)
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

func (f *Fs) openContent(owner, repository, branch, name string) (*Content, error) {
	return &Content{
		owner:      owner,
		repository: repository,
		branch:     branch,
		name:       name,
	}, nil
}

func (f *Fs) openRelease(owner, repository, name string) (*Release, error) {
	return &Release{
		owner:      owner,
		repository: repository,
		name:       name,
	}, nil
}

func (f *Fs) openAsset(owner, repository, release, name string) (*Asset, error) {
	return &Asset{
		owner:      owner,
		repository: repository,
		release:    release,
		name:       name,
	}, nil
}

func background(*Fs, ihfs.Operation) context.Context {
	return context.Background()
}

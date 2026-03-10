package ghfs

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/google/go-github/v84/github"
	"github.com/unmango/go/fopt"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/op"
)

var errNotImplemented = fmt.Errorf("github: %w", ihfs.ErrNotImplemented)

type ContextFunc func(*Fs, ihfs.Operation) context.Context

type Fs struct {
	client *github.Client
	token  string
	ctxFn  ContextFunc
}

func New(options ...Option) *Fs {
	f := &Fs{ctxFn: background}
	fopt.ApplyAll(f, options)
	if f.client == nil {
		f.client = github.NewClient(nil)
	}
	if f.token != "" {
		f.client = f.client.WithAuthToken(f.token)
	}

	return f
}

func (*Fs) Name() string {
	return "github"
}

func (f *Fs) Open(name string) (ihfs.File, error) {
	if name == "." {
		return &File{name: ".", isDir: true}, nil
	}
	if !strings.Contains(name, "://") && !fs.ValidPath(name) {
		return nil, openErr(name, ihfs.ErrInvalid)
	}
	return f.open(name)
}

func (f *Fs) context(op ihfs.Operation) context.Context {
	return f.ctxFn(f, op)
}

func Open(fsys ihfs.FS, name string) (*File, error) {
	if fs, ok := fsys.(*Fs); ok {
		return fs.open(name)
	}
	return nil, openErr(name, errNotImplemented)
}

func (f *Fs) open(name string) (*File, error) {
	path, err := Parse(name)
	if err != nil {
		return nil, openErr(name, err)
	}

	ctx := f.context(op.Open{Name: path.Name()})
	if id, err := f.assetId(ctx, path); err == nil && id != 0 {
		return open(ctx, f.client, assetPath(path.owner, path.repo, id))
	} else if path.asset != "" && err != nil {
		return nil, openErr(name, err)
	}

	return open(ctx, f.client, path.APIPath())
}

func (f *Fs) assetId(ctx context.Context, p Path) (int64, error) {
	return assetId(ctx, f.client, p)
}

func background(*Fs, ihfs.Operation) context.Context {
	return context.Background()
}

func release(ctx context.Context, c *github.Client, p Path) (*github.RepositoryRelease, error) {
	if p.releaseID != 0 {
		r, _, err := c.Repositories.GetRelease(ctx,
			p.owner, p.repo, p.releaseID,
		)
		return r, err
	}
	if p.tag != "" {
		r, _, err := c.Repositories.GetReleaseByTag(ctx,
			p.owner, p.repo, p.tag,
		)
		return r, err
	}

	return nil, fmt.Errorf("release not specified")
}

func assetId(ctx context.Context, c *github.Client, p Path) (int64, error) {
	if p.assetID != 0 {
		return p.assetID, nil
	}
	if p.asset == "" {
		return 0, fmt.Errorf("empty asset name")
	}

	rel, err := release(ctx, c, p)
	if err != nil {
		return 0, err
	}
	for _, asset := range rel.Assets {
		if asset.GetName() == p.asset {
			return asset.GetID(), nil
		}
	}

	return 0, ihfs.ErrNotExist
}

func open(ctx context.Context, c *github.Client, url string) (*File, error) {
	if r, err := do(ctx, c, url); err != nil {
		return nil, err
	} else {
		return &File{name: url, rc: r}, nil
	}
}

func do(ctx context.Context, c *github.Client, url string) (io.ReadCloser, error) {
	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.BareDo(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func openErr(name string, err error) error {
	return &ihfs.PathError{
		Op:   "open",
		Path: name,
		Err:  err,
	}
}

package ghfs

import (
	"context"
	"encoding/json"
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
		return &Dir{name: "."}, nil
	}
	if !strings.Contains(name, "://") && !fs.ValidPath(name) {
		return nil, openErr(name, ihfs.ErrInvalid)
	}
	return f.open(name)
}

func (f *Fs) context(op ihfs.Operation) context.Context {
	return f.ctxFn(f, op)
}

func (f *Fs) do(ctx context.Context, url string) (io.ReadCloser, error) {
	return do(ctx, f.client, url)
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
	if path.Asset() != "" {
		return f.openAssetByName(ctx, path)
	}

	r, err := f.do(ctx, path.APIPath())
	if err != nil {
		return nil, openErr(path.String(), err)
	}

	return &File{r, path.String()}, nil
}

func (f *Fs) openAssetByName(ctx context.Context, p *Path) (*File, error) {
	releaseBody, err := f.do(ctx, releasePath())
	if err != nil {
		return nil, openErr(p.String(), err)
	}
	defer releaseBody.Close()

	var release github.RepositoryRelease
	if err := json.NewDecoder(releaseBody).Decode(&release); err != nil {
		return nil, openErr(p.String(), err)
	}

	for _, asset := range release.Assets {
		if asset.GetName() == p.Asset() {
			return OpenAsset(f)
		}
	}

	return nil, openErr(p.String(), ihfs.ErrNotExist)
}

func background(*Fs, ihfs.Operation) context.Context {
	return context.Background()
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

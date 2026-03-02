package ghfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"strings"

	"github.com/google/go-github/v82/github"
	"github.com/unmango/go/fopt"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/op"
)

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
		return openErr(name, ihfs.ErrInvalid)
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

	return openErr(name, fmt.Errorf("github: %w", ihfs.ErrNotImplemented))
}

func (f *Fs) open(name string) (*File, error) {
	path, err := normalize(name)
	if err != nil {
		if _, ok := err.(*ihfs.PathError); ok {
			return nil, err
		}
		return openErr(name, err)
	}

	if rest, ok := strings.CutPrefix(path, assetLookupPrefix); ok {
		parts := strings.SplitN(rest, "/", 4)
		if len(parts) == 4 {
			tag, _ := url.PathUnescape(parts[2])
			assetName, _ := url.PathUnescape(parts[3])
			return f.openAssetByName(name, parts[0], parts[1], tag, assetName)
		}
	}

	ctx := f.context(op.Open{Name: name})
	r, err := f.do(ctx, path)
	if err != nil {
		return openErr(name, err)
	}
	return &File{r, name}, nil
}

func (f *Fs) openAssetByName(name, owner, repo, tag, assetName string) (*File, error) {
	ctx := f.context(op.Open{Name: name})

	releaseBody, err := f.do(ctx, releasePath(owner, repo, tag))
	if err != nil {
		return openErr(name, err)
	}
	defer releaseBody.Close()

	var release github.RepositoryRelease
	if err := json.NewDecoder(releaseBody).Decode(&release); err != nil {
		return openErr(name, err)
	}

	for _, asset := range release.Assets {
		if asset.GetName() == assetName {
			assetURL := fmt.Sprintf("repos/%v/%v/releases/assets/%v", owner, repo, asset.GetID())
			r, err := f.do(ctx, assetURL)
			if err != nil {
				return openErr(name, err)
			}
			return &File{r, name}, nil
		}
	}

	return openErr(name, ihfs.ErrNotExist)
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

func openErr(name string, err error) (*File, error) {
	return nil, &ihfs.PathError{
		Op:   "open",
		Path: name,
		Err:  err,
	}
}

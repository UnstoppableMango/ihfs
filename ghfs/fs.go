package ghfs

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/go-github/v82/github"
	"github.com/unmango/go/fopt"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/op"
)

type ContextFunc func(*Fs, ihfs.Operation) context.Context

type Fs struct {
	client *github.Client
	ctxFn  ContextFunc
}

func New(options ...Option) *Fs {
	f := &Fs{ctxFn: background}
	fopt.ApplyAll(f, options)
	if f.client == nil {
		f.client = github.NewClient(nil)
	}

	return f
}

func (*Fs) Name() string {
	return "github"
}

func (f *Fs) Open(name string) (ihfs.File, error) {
	return f.open(name)
}

func (f *Fs) setAuthToken(token string) {
	f.client = f.client.WithAuthToken(token)
}

func (f *Fs) context(op ihfs.Operation) context.Context {
	return f.ctxFn(f, op)
}

func (f *Fs) do(ctx context.Context, url string) (*bytes.Reader, error) {
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
		return openErr(name, err)
	}

	ctx := f.context(op.Open{Name: name})
	if r, err := f.do(ctx, path); err != nil {
		return openErr(name, err)
	} else {
		return &File{r, name}, nil
	}
}

func background(*Fs, ihfs.Operation) context.Context {
	return context.Background()
}

func do(ctx context.Context, c *github.Client, url string) (*bytes.Reader, error) {
	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	_, err = c.Do(ctx, req, buf)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}

func openErr(name string, err error) (*File, error) {
	return nil, &ihfs.PathError{
		Op:   "open",
		Path: name,
		Err:  err,
	}
}

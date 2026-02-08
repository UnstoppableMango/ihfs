package ghfs

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"time"

	"github.com/google/go-github/v82/github"
	"github.com/unstoppablemango/ihfs"
)

type file struct {
	*bytes.Reader
	name string
}

func (f *file) IsDir() bool                  { return false }
func (f *file) ModTime() time.Time           { return time.Time{} }
func (f *file) Mode() fs.FileMode            { return 0444 } // read-only permissions (regular file)
func (f *file) Size() int64                  { return int64(f.Len()) }
func (f *file) Sys() any                     { return f.Reader }
func (f *file) Close() error                 { return nil }
func (f *file) Name() string                 { return f.name }
func (f *file) Stat() (ihfs.FileInfo, error) { return f, nil }

func (f *file) dec(v any) error {
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	return json.NewDecoder(f.Reader).Decode(v)
}

type Owner struct{ *file }

func (o *Owner) User() (*github.User, error) {
	var user github.User
	if err := o.dec(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

type Repository struct {
	*file
	owner string
}

func (r *Repository) Owner() string { return r.owner }

func (r *Repository) Repository() (*github.Repository, error) {
	var repo github.Repository
	if err := r.dec(&repo); err != nil {
		return nil, err
	}

	return &repo, nil
}

type Release struct {
	*file
	owner      string
	repository string
}

func (r *Release) Owner() string      { return r.owner }
func (r *Release) Repository() string { return r.repository }

func (r *Release) Release() (*github.RepositoryRelease, error) {
	var release github.RepositoryRelease
	if err := r.dec(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

type Asset struct {
	*file
	owner      string
	repository string
	release    string
}

func (a *Asset) Release() string    { return a.release }
func (a *Asset) Repository() string { return a.repository }
func (a *Asset) Owner() string      { return a.owner }

func (a *Asset) Asset() (*github.ReleaseAsset, error) {
	var asset github.ReleaseAsset
	if err := a.dec(&asset); err != nil {
		return nil, err
	}

	return &asset, nil
}

type Branch struct {
	*file
	owner      string
	repository string
}

func (b *Branch) Owner() string      { return b.owner }
func (b *Branch) Repository() string { return b.repository }

func (b *Branch) Branch() (*github.Branch, error) {
	var branch github.Branch
	if err := b.dec(&branch); err != nil {
		return nil, err
	}

	return &branch, nil
}

type Content struct {
	*file
	owner      string
	repository string
	branch     string
}

func (c *Content) Owner() string      { return c.owner }
func (c *Content) Repository() string { return c.repository }
func (c *Content) Branch() string     { return c.branch }

func (c *Content) Content() (*github.RepositoryContent, error) {
	var content github.RepositoryContent
	if err := c.dec(&content); err != nil {
		return nil, err
	}

	return &content, nil
}

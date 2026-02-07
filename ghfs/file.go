package ghfs

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/google/go-github/v82/github"
	"github.com/unstoppablemango/ihfs"
)

type Owner struct {
	name string
	r    *bytes.Reader
}

// Close implements [fs.File].
func (*Owner) Close() error   { return nil }
func (o *Owner) Name() string { return o.name }

// Read implements [fs.File].
func (o *Owner) Read(p []byte) (n int, err error) {
	return o.r.Read(p)
}

// Stat implements [fs.File].
func (o *Owner) Stat() (ihfs.FileInfo, error) {
	return nil, nil
}

func (o *Owner) User() (*github.User, error) {
	var user github.User
	if err := dec(o.r, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

type Repository struct {
	name  string
	owner string
	r     *bytes.Reader
}

// Close implements [fs.File].
func (r *Repository) Close() error  { return nil }
func (r *Repository) Name() string  { return r.name }
func (r *Repository) Owner() string { return r.owner }

// Read implements [fs.File].
func (r *Repository) Read([]byte) (int, error) {
	return 0, nil
}

// Stat implements [fs.File].
func (r *Repository) Stat() (ihfs.FileInfo, error) {
	return nil, nil
}

func (r *Repository) Repository() (*github.Repository, error) {
	var repo github.Repository
	if err := dec(r.r, &repo); err != nil {
		return nil, err
	}

	return &repo, nil
}

type Release struct {
	name       string
	owner      string
	repository string
	r          *bytes.Reader
}

// Close implements [fs.File].
func (r *Release) Close() error       { return nil }
func (r *Release) Name() string       { return r.name }
func (r *Release) Owner() string      { return r.owner }
func (r *Release) Repository() string { return r.repository }

// Read implements [fs.File].
func (r *Release) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

// Stat implements [fs.File].
func (r *Release) Stat() (ihfs.FileInfo, error) {
	return nil, nil
}

func (r *Release) Release() (*github.RepositoryRelease, error) {
	var release github.RepositoryRelease
	if err := dec(r.r, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

type Asset struct {
	name       string
	owner      string
	repository string
	release    string
	r          *bytes.Reader
}

// Close implements [fs.File].
func (a *Asset) Close() error       { return nil }
func (a *Asset) Name() string       { return a.name }
func (a *Asset) Release() string    { return a.release }
func (a *Asset) Repository() string { return a.repository }
func (a *Asset) Owner() string      { return a.owner }

// Read implements [fs.File].
func (a *Asset) Read(p []byte) (int, error) {
	return a.r.Read(p)
}

// Stat implements [fs.File].
func (a *Asset) Stat() (ihfs.FileInfo, error) {
	return nil, nil
}

func (a *Asset) Asset() (*github.ReleaseAsset, error) {
	var asset github.ReleaseAsset
	if err := dec(a.r, &asset); err != nil {
		return nil, err
	}

	return &asset, nil
}

type Branch struct {
	name       string
	owner      string
	repository string
	r          *bytes.Reader
}

// Close implements [fs.File].
func (b *Branch) Close() error       { return nil }
func (b *Branch) Name() string       { return b.name }
func (b *Branch) Owner() string      { return b.owner }
func (b *Branch) Repository() string { return b.repository }

// Read implements [fs.File].
func (b *Branch) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

// Stat implements [fs.File].
func (b *Branch) Stat() (ihfs.FileInfo, error) {
	return nil, nil
}

func (b *Branch) Branch() (*github.Branch, error) {
	var branch github.Branch
	if err := dec(b.r, &branch); err != nil {
		return nil, err
	}

	return &branch, nil
}

type Content struct {
	name       string
	owner      string
	repository string
	branch     string
	r          *bytes.Reader
}

// Close implements [fs.File].
func (c *Content) Close() error       { return nil }
func (c *Content) Name() string       { return c.name }
func (c *Content) Owner() string      { return c.owner }
func (c *Content) Repository() string { return c.repository }
func (c *Content) Branch() string     { return c.branch }

// Read implements [fs.File].
func (c *Content) Read(p []byte) (int, error) {
	return c.r.Read(p)
}

// Stat implements [fs.File].
func (c *Content) Stat() (ihfs.FileInfo, error) {
	return nil, nil
}

func (c *Content) Content() (*github.RepositoryContent, error) {
	var content github.RepositoryContent
	if err := dec(c.r, &content); err != nil {
		return nil, err
	}

	return &content, nil
}

func dec(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

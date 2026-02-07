package ghfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/go-github/v82/github"
	"github.com/unmango/go/fopt"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/op"
)

type (
	ContextFunc func(*Fs, ihfs.Operation) context.Context
)

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
	cleaned := clean(name)
	if cleaned == "" {
		return nil, &ihfs.PathError{
			Op:   "open",
			Path: name,
			Err:  ihfs.ErrInvalid,
		}
	}

	parts := strings.Split(cleaned, "/")

	// TODO: API path patterns
	// will likely need to use the URL prefix to determine which pattern to use
	// also, potential to simply pass the given path directly w/o cleaning
	switch len(parts) {
	case 1:
		return f.openOwner(parts[0])
	case 2:
		return f.openRepository(parts[0], parts[1])
	case 4:
		// Expected pattern: owner/repo/tree/branch
		if parts[2] == "tree" {
			return f.openBranch(parts[0], parts[1], parts[3])
		}
	case 5:
		// Expected patterns:
		// - owner/repo/blob/branch/path
		// - owner/repo/releases/(tag|download)/TAG
		switch parts[2] {
		case "blob":
			return f.openContent(parts[0], parts[1], parts[3], parts[4])
		case "releases":
			if parts[3] == "tag" || parts[3] == "download" {
				return f.openRelease(parts[0], parts[1], parts[4])
			}
		}
	}

	if len(parts) >= 6 {
		// Expected patterns:
		// - owner/repo/releases/(tag|download)/TAG/asset
		// - owner/repo/(tree|blob)/branch/path/to/item
		if parts[2] == "releases" && (parts[3] == "tag" || parts[3] == "download") {
			return f.openAsset(parts[0], parts[1], parts[4], parts[5])
		}

		if parts[2] == "tree" || parts[2] == "blob" {
			return f.openContent(parts[0], parts[1], parts[3],
				strings.Join(parts[4:], "/"),
			)
		}
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

func (f *Fs) do(ctx context.Context, url string) (*bytes.Reader, error) {
	return do(ctx, f.client, url)
}

func (f *Fs) open(name, url string) (*file, error) {
	r, err := f.do(f.context(op.Open{Name: name}), url)
	if err != nil {
		return nil, err
	}

	return &file{
		name:   name,
		Reader: r,
	}, nil
}

func (f *Fs) openOwner(name string) (*Owner, error) {
	file, err := f.open(name, fmt.Sprintf("users/%v", name))
	if err != nil {
		return nil, err
	}

	return &Owner{file: file}, nil
}

func (f *Fs) openRepository(owner, name string) (*Repository, error) {
	url := fmt.Sprintf("repos/%v/%v", owner, name)
	file, err := f.open(name, url)
	if err != nil {
		return nil, err
	}

	return &Repository{
		file:  file,
		owner: owner,
	}, nil
}

func (f *Fs) openBranch(owner, repository, name string) (*Branch, error) {
	url := fmt.Sprintf("repos/%v/%v/branches/%v", owner, repository, name)
	file, err := f.open(name, url)
	if err != nil {
		return nil, err
	}

	return &Branch{
		file:       file,
		owner:      owner,
		repository: repository,
	}, nil
}

func (f *Fs) openContent(owner, repository, branch, name string) (*Content, error) {
	// Escape each path segment individually to preserve forward slashes
	pathSegments := strings.Split(name, "/")
	escapedSegments := make([]string, len(pathSegments))
	for i, segment := range pathSegments {
		escapedSegments[i] = url.PathEscape(segment)
	}
	escapedPath := strings.Join(escapedSegments, "/")
	escapedRef := url.QueryEscape(branch)
	apiURL := fmt.Sprintf("repos/%v/%v/contents/%v?ref=%v", owner, repository, escapedPath, escapedRef)
	file, err := f.open(name, apiURL)
	if err != nil {
		return nil, err
	}

	return &Content{
		file:       file,
		owner:      owner,
		repository: repository,
		branch:     branch,
	}, nil
}

func (f *Fs) openRelease(owner, repository, name string) (*Release, error) {
	url := fmt.Sprintf("repos/%v/%v/releases/tags/%v", owner, repository, name)
	file, err := f.open(name, url)
	if err != nil {
		return nil, err
	}

	return &Release{
		file:       file,
		owner:      owner,
		repository: repository,
	}, nil
}

func (f *Fs) openAsset(owner, repository, releaseTag, assetName string) (*Asset, error) {
	// First, fetch the release to get its assets
	releaseURL := fmt.Sprintf("repos/%v/%v/releases/tags/%v", owner, repository, releaseTag)
	releaseBody, err := f.do(f.context(op.Open{Name: releaseTag}), releaseURL)
	if err != nil {
		return nil, err
	}

	// Decode the release to get the assets list
	var release github.RepositoryRelease
	if err := json.NewDecoder(releaseBody).Decode(&release); err != nil {
		return nil, err
	}

	// Find the asset by name
	var targetAsset *github.ReleaseAsset
	for _, asset := range release.Assets {
		if asset.Name != nil && *asset.Name == assetName {
			targetAsset = asset
			break
		}
	}

	if targetAsset == nil {
		return nil, &ihfs.PathError{
			Op:   "open",
			Path: assetName,
			Err:  ihfs.ErrNotExist,
		}
	}

	// Encode the asset metadata to JSON
	assetJSON, err := json.Marshal(targetAsset)
	if err != nil {
		return nil, err
	}

	return &Asset{
		file: &file{
			name:   assetName,
			Reader: bytes.NewReader(assetJSON),
		},
		owner:      owner,
		repository: repository,
		release:    releaseTag,
	}, nil
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

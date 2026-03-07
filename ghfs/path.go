package ghfs

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/unstoppablemango/ihfs"
)

const assetLookupPrefix = "ghfs-asset:"

var hosts = []string{
	"github.com",
	"api.github.com",
	"raw.githubusercontent.com",
}

type Path struct {
	name    string
	u       *url.URL
	host    string
	apiPath string

	owner   string
	repo    string
	branch  string
	tag     string
	asset   string
	release string
	content string
}

func (p *Path) String() string  { return p.name }
func (p *Path) Host() string    { return p.host }
func (p *Path) APIPath() string { return p.apiPath }
func (p *Path) Owner() string   { return p.owner }
func (p *Path) Repo() string    { return p.repo }
func (p *Path) Branch() string  { return p.branch }
func (p *Path) Tag() string     { return p.tag }
func (p *Path) Asset() string   { return p.asset }
func (p *Path) Release() string { return p.release }
func (p *Path) Content() string { return p.content }

func Parse(name string) (p *Path, err error) {
	p = &Path{name: name}
	if p.u, err = url.Parse(name); err != nil {
		return nil, err
	}

	p.host, p.apiPath = p.splitHost()
	switch p.host {
	case "api.github.com":
		// For API URLs, the path is already in the correct format
	case "github.com":
		if err = p.asWebURL(); err != nil {
			return nil, err
		}
	case "raw.githubusercontent.com":
		if err = p.asRawURL(); err != nil {
			return nil, err
		}
	case "":
		p.apiPath = p.u.RequestURI()
	default:
		return nil, fmt.Errorf("invalid host: %s", p.host)
	}

	return p, nil
}

func (p *Path) splitHost() (host, rest string) {
	if host = p.u.Hostname(); host != "" {
		return host, p.u.RequestURI()
	}

	for _, h := range hosts {
		if after, ok := strings.CutPrefix(p.name, h); ok {
			return h, strings.TrimPrefix(after, "/")
		}
	}

	return "", p.name
}

func (p *Path) asWebURL() error {
	// TODO: Refactor to avoid mutation
	parts := strings.Split(clean(p.name), "/")
	for i, p := range parts {
		if unescaped, err := url.PathUnescape(p); err == nil {
			parts[i] = unescaped
		}
	}

	switch len(parts) {
	case 1:
		p.owner = parts[0]
		return nil
	case 2:
		p.owner = parts[0]
		p.repo = parts[1]
		return nil
	case 4:
		// Expected pattern: owner/repo/tree/branch
		if parts[2] == "tree" {
			p.owner = parts[0]
			p.repo = parts[1]
			p.branch = parts[3]
			return nil
		}
	case 5:
		// Expected patterns:
		// - owner/repo/blob/branch/path
		// - owner/repo/tree/branch/path
		// - owner/repo/releases/(tag|download)/TAG
		switch parts[2] {
		case "blob", "tree":
			p.owner = parts[0]
			p.repo = parts[1]
			p.branch = parts[3]
			p.content = parts[4]
			return nil
		case "releases":
			if parts[3] == "tag" || parts[3] == "download" {
				p.owner = parts[0]
				p.repo = parts[1]
				p.tag = parts[4]
				return nil
			}
		}
	}

	if len(parts) >= 6 {
		// Expected patterns:
		// - owner/repo/releases/(tag|download)/TAG/asset
		// - owner/repo/(tree|blob)/branch/path/to/item
		if parts[2] == "releases" && (parts[3] == "tag" || parts[3] == "download") {
			p.owner = parts[0]
			p.repo = parts[1]
			p.tag = parts[4]
			p.asset = parts[5]
			return nil
		}

		if parts[2] == "tree" || parts[2] == "blob" {
			p.owner = parts[0]
			p.repo = parts[1]
			p.branch = parts[3]
			p.content = strings.Join(parts[4:], "/")
			return nil
		}
	}

	return &ihfs.PathError{
		Op:   "open",
		Path: p.name,
		Err:  ihfs.ErrNotExist,
	}
}

func (p *Path) asRawURL() error {
	// TODO: Refactor to avoid mutation
	parts := strings.Split(clean(p.name), "/")
	for i, p := range parts {
		if unescaped, err := url.PathUnescape(p); err == nil {
			parts[i] = unescaped
		}
	}

	switch len(parts) {
	case 1:
		p.owner = parts[0]
	case 2:
		p.owner = parts[0]
		p.repo = parts[1]
	case 3:
		p.owner = parts[0]
		p.repo = parts[1]
		p.branch = parts[2]
	default:
		p.owner = parts[0]
		p.repo = parts[1]
		p.branch = parts[2]
		p.content = strings.Join(parts[3:], "/")
	}

	return nil
}

func clean(path string) string {
	if u, err := url.Parse(path); err == nil {
		path = u.EscapedPath()
	}
	for _, h := range hosts {
		path = strings.TrimPrefix(path, h)
	}

	return strings.TrimLeft(path, "/")
}

func (p *Path) ownerPath() string {
	return fmt.Sprintf("users/%v", p.owner)
}

func (p *Path) repoPath() string {
	return fmt.Sprintf("repos/%v/%v", p.owner, p.repo)
}

func (p *Path) branchPath() string {
	return fmt.Sprintf("repos/%v/%v/branches/%v", p.owner, p.repo, url.PathEscape(p.branch))
}

func (p *Path) contentPath() string {
	segments := strings.Split(p.content, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}

	return fmt.Sprintf(
		"repos/%v/%v/contents/%v?ref=%v",
		p.owner,
		p.repo,
		strings.Join(segments, "/"),
		url.QueryEscape(p.branch),
	)
}

func (p *Path) releasePath() string {
	return fmt.Sprintf(
		"repos/%v/%v/releases/tags/%v",
		p.owner, p.repo, url.PathEscape(p.tag),
	)
}

func (p *Path) assetPath() string {
	return fmt.Sprintf(
		"%v/%v/%v/%v",
		p.owner,
		p.repo,
		url.PathEscape(p.tag),
		url.PathEscape(p.asset),
	)
}

package ghfs

import (
	"fmt"
	"net/url"
	"strings"
)

var hosts = []string{
	"github.com",
	"api.github.com",
	"raw.githubusercontent.com",
}

type Path struct {
	name string
	u    *url.URL
	host string
	path string

	owner   string
	repo    string
	branch  string
	tag     string
	release string
	asset   string
	content []string
}

func (p Path) Name() string    { return p.name }
func (p Path) String() string  { return p.name }
func (p Path) Owner() string   { return p.owner }
func (p Path) Repo() string    { return p.repo }
func (p Path) Branch() string  { return p.branch }
func (p Path) Tag() string     { return p.tag }
func (p Path) Asset() string   { return p.asset }
func (p Path) Release() string { return p.release }

func (p *Path) Host() string {
	return p.host
}

func Parse(name string) (p Path, err error) {
	p.name = name
	if p.u, err = url.Parse(name); err != nil {
		return Path{}, err
	}

	p.host, p.path = splitHost(name, p.u)
	parts := strings.Split(strings.TrimLeft(p.path, "/"), "/")
	switch p.host {
	case "github.com":
		asWeb(&p, parts)
	case "api.github.com":
	// TODO
	case "raw.githubusercontent.com":
		asRaw(&p, parts)
	}

	return p, nil
}

func splitHost(name string, u *url.URL) (host, rest string) {
	if host = u.Hostname(); host != "" {
		return host, u.RequestURI()
	}

	for _, h := range hosts {
		if after, ok := strings.CutPrefix(name, h); ok {
			return h, strings.TrimPrefix(after, "/")
		}
	}

	return "", name
}

func (p Path) APIPath() string {
	return p.path
}

func (p Path) releasePath() string {
	return releasePath(p.owner, p.repo, p.tag)
}

func asWeb(p *Path, parts []string) {
	for i, s := range parts {
		switch i {
		case 0:
			p.owner = s
		case 1:
			p.repo = s
		case 3:
			if parts[2] == "tree" {
				p.branch = s
			}
		case 5:
			switch parts[2] {
			case "blob", "tree":
				p.content = append(p.content, s)
			case "releases":
				if parts[3] == "tag" || parts[3] == "download" {
					p.release = s
				}
			}
		}

		if i >= 6 {
			switch parts[2] {
			case "blob", "tree":
				p.content = append(p.content, s)
			case "releases":
				if parts[3] == "tag" || parts[3] == "download" {
					p.release = parts[5]
				}
			}
		}
	}
}

func asRaw(p *Path, parts []string) {
	if len(parts) > 0 {
		p.owner = parts[0]
	}
	if len(parts) > 1 {
		p.repo = parts[1]
	}
	if len(parts) > 2 {
		p.branch = parts[2]
	}
	if len(parts) > 3 {
		p.content = parts[3:]
	}
}

func ownerPath(owner string) string {
	return fmt.Sprintf("users/%v", owner)
}

func repoPath(owner, repo string) string {
	return fmt.Sprintf("repos/%v/%v", owner, repo)
}

func branchPath(owner, repo, branch string) string {
	return fmt.Sprintf("repos/%v/%v/branches/%v", owner, repo, branch)
}

func contentPath(owner, repo, branch, content string) string {
	return fmt.Sprintf(
		"repos/%v/%v/contents/%v?ref=%v",
		owner, repo, content, branch,
	)
}

func releasePath(owner, repo, tag string) string {
	return fmt.Sprintf(
		"repos/%v/%v/releases/tags/%v",
		owner, repo, url.PathEscape(tag),
	)
}

func assetPath(owner, repo string, id int64) string {
	return fmt.Sprintf("repos/%v/%v/releases/assets/%v", owner, repo, id)
}

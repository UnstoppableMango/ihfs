package ghfs

import (
	"fmt"
	"net/url"
	"slices"
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
}

func (p *Path) Name() string   { return p.name }
func (p *Path) String() string { return p.name }

// func (p *Path) Owner() string   { return p.owner }
// func (p *Path) Repo() string    { return p.repo }
// func (p *Path) Branch() string  { return p.branch }
// func (p *Path) Tag() string     { return p.tag }
// func (p *Path) Asset() string   { return p.asset }
// func (p *Path) Release() string { return p.release }
// func (p *Path) Content() string { return p.content }

func (p *Path) Host() string {
	return p.host
}

func Parse(name string) (*Path, error) {
	u, err := url.Parse(name)
	if err != nil {
		return nil, err
	}

	host, path := splitHost(name, u)
	if !slices.Contains(append(hosts, ""), host) {
		return nil, fmt.Errorf("invalid host: %s", host)
	}

	return &Path{
		name: name,
		u:    u,
		host: host,
		path: path,
	}, nil
}

func (p *Path) parts() parts {
	return strings.Split(strings.TrimLeft(p.path, "/"), "/")
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

func (p *Path) APIPath() string {
	return p.path
}

type parts []string

func (p parts) fromWeb() (path, asset string) {
	switch len(p) {
	case 1:
		return ownerPath(p[0]), ""
	case 2:
		return repoPath(p[0], p[1]), ""
	case 4:
		if p[2] == "tree" {
			return branchPath(p[0], p[1], p[3]), ""
		}
	case 5:
		switch p[2] {
		case "blob", "tree":
			return contentPath(p[0], p[1], p[3], p[4]), ""
		case "releases":
			if p[3] == "tag" || p[3] == "download" {
				return releasePath(p[0], p[1], p[4]), ""
			}
		}
	case 6:
		switch p[2] {
		case "blob", "tree":
			return contentPath(p[0], p[1], p[3], strings.Join(p[4:], "/")), ""
		case "releases":
			if p[3] == "tag" || p[3] == "download" {
				return releasePath(p[0], p[1], p[4]), p[5]
			}
		}
	}

	return "", ""
}

func fromWebPath(name string) (path, asset string) {
	return parts(strings.Split(strings.TrimLeft(name, "/"), "/")).fromWeb()
}

// func (p *Path) applyWebPath(name string) {
// 	parts := strings.Split(strings.TrimLeft(name, "/"), "/")
// 	for i, s := range parts {
// 		if unescaped, err := url.PathUnescape(s); err == nil {
// 			parts[i] = unescaped
// 		}
// 	}

// 	if len(parts) > 0 {
// 		p.owner = parts[0]
// 	}
// 	if len(parts) > 1 {
// 		p.repo = parts[1]
// 	}
// 	if len(parts) > 3 && parts[2] == "tree" {
// 		p.branch = parts[3]
// 	}
// 	if len(parts) > 4 {
// 		switch parts[2] {
// 		case "blob", "tree":
// 			p.content = parts[4]
// 		case "releases":
// 			if parts[3] == "tag" || parts[3] == "download" {
// 				p.tag = parts[4]
// 			}
// 		}
// 	}

// 	if len(parts) >= 6 {
// 		switch parts[2] {
// 		case "blob", "tree":
// 			p.content = strings.Join(parts[4:], "/")
// 		case "releases":
// 			if parts[3] == "tag" || parts[3] == "download" {
// 				p.asset = parts[5]
// 			}
// 		}
// 	}
// }

// func (p *Path) applyRawPath(url string) {
// 	parts := strings.Split(strings.TrimLeft(url, "/"), "/")
// 	if len(parts) > 0 {
// 		p.owner = parts[0]
// 	}
// 	if len(parts) > 1 {
// 		p.repo = parts[1]
// 	}
// 	if len(parts) > 2 {
// 		p.branch = parts[2]
// 	}
// 	if len(parts) > 3 {
// 		p.content = strings.Join(parts[3:], "/")
// 	}
// }

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

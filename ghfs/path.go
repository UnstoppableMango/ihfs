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

	owner   string
	repo    string
	branch  string
	tag     string
	release string
	asset   string
	content []string
}

func Parse(name string) (p Path, err error) {
	p.name = name
	if p.u, err = url.Parse(name); err != nil {
		return Path{}, err
	}

	var path string
	p.host, path = splitHost(name, p.u)
	pathOnly, _, _ := strings.Cut(path, "?")
	parts := strings.Split(strings.TrimLeft(pathOnly, "/"), "/")
	switch p.host {
	case "github.com":
		asWeb(&p, parts)
	case "api.github.com", "":
		asAPI(&p, parts)
	case "raw.githubusercontent.com":
		asRaw(&p, parts)
	default:
		return Path{}, fmt.Errorf("invalid host: %s", p.host)
	}

	return p, nil
}

func (p Path) Name() string    { return p.name }
func (p Path) String() string  { return p.name }
func (p Path) Host() string    { return p.host }
func (p Path) Owner() string   { return p.owner }
func (p Path) Repo() string    { return p.repo }
func (p Path) Branch() string  { return p.branch }
func (p Path) Tag() string     { return p.tag }
func (p Path) Asset() string   { return p.asset }
func (p Path) Release() string { return p.release }

func (p Path) APIPath() string {
	prefix := ""
	if p.host == "api.github.com" && p.u != nil && p.u.Scheme != "" {
		prefix = "/"
	}

	if p.owner == "" {
		if p.host == "github.com" || p.host == "raw.githubusercontent.com" {
			return "user"
		}
		return prefix
	}
	if p.repo == "" {
		return prefix + ownerPath(p.owner)
	}
	if p.release != "" {
		return prefix + releasePath(p.owner, p.repo, p.tag)
	}
	if len(p.content) > 0 {
		return prefix + contentPath(p.owner, p.repo, p.branch, strings.Join(p.content, "/"))
	}
	if p.branch != "" {
		return prefix + branchPath(p.owner, p.repo, p.branch)
	}
	return prefix + repoPath(p.owner, p.repo)
}

func splitHost(name string, u *url.URL) (host, rest string) {
	if host = u.Hostname(); host != "" {
		return host, u.RequestURI()
	}

	for _, h := range hosts {
		if after, ok := strings.CutPrefix(name, h); ok {
			p, _, _ := strings.Cut(strings.TrimLeft(after, "/"), "?")
			return h, p
		}
	}

	return "", name
}

func asWeb(p *Path, parts []string) {
	if len(parts) > 0 {
		p.owner = parts[0]
	}
	if len(parts) > 1 {
		p.repo = parts[1]
	}

	if len(parts) < 3 {
		return
	}

	switch parts[2] {
	case "tree":
		if len(parts) > 3 {
			p.branch = parts[3]
		}
		if len(parts) > 4 {
			p.content = parts[4:]
		}
	case "blob":
		if len(parts) > 3 {
			p.branch = parts[3]
		}
		if len(parts) > 4 {
			p.content = parts[4:]
		}
	case "releases":
		if len(parts) > 3 && (parts[3] == "tag" || parts[3] == "download") {
			if len(parts) > 4 {
				p.tag = parts[4]
				p.release = parts[4]
			}
			if len(parts) > 5 {
				p.asset = parts[5]
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

func asAPI(p *Path, parts []string) {
	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
		return
	}

	switch parts[0] {
	case "users":
		if len(parts) > 1 {
			p.owner = parts[1]
		}
	case "repos":
		if len(parts) > 1 {
			p.owner = parts[1]
		}
		if len(parts) > 2 {
			p.repo = parts[2]
		}
		if len(parts) > 3 {
			switch parts[3] {
			case "branches":
				if len(parts) > 4 {
					p.branch = parts[4]
				}
			case "releases":
				if len(parts) > 4 {
					switch parts[4] {
					case "tags":
						if len(parts) > 5 {
							p.tag = parts[5]
							p.release = parts[5]
						}
					}
				}
			case "contents":
				if len(parts) > 4 {
					p.content = parts[4:]
				}
				if p.u != nil {
					p.branch = p.u.Query().Get("ref")
				}
			}
		}
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

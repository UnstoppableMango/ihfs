package ghfs

import (
	"fmt"
	"net/url"
	"strconv"
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
	content []string

	tag       string
	releaseID int64

	asset   string
	assetID int64
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
	case "api.github.com", "":
		asAPI(&p, parts)
	case "github.com":
		asWeb(&p, parts)
	case "raw.githubusercontent.com":
		asRaw(&p, parts)
	default:
		return Path{}, fmt.Errorf("invalid host: %s", p.host)
	}

	return p, nil
}

func (p Path) Name() string   { return p.name }
func (p Path) String() string { return p.name }
func (p Path) Host() string   { return p.host }
func (p Path) Owner() string  { return p.owner }
func (p Path) Repo() string   { return p.repo }
func (p Path) Branch() string { return p.branch }
func (p Path) Tag() string    { return p.tag }
func (p Path) Asset() string  { return p.asset }
func (p Path) Release() int64 { return p.releaseID }

func (p Path) APIPath() string {
	if p.owner == "" {
		return "user"
	}
	if p.repo == "" {
		return ownerPath(p.owner)
	}
	if p.releaseID != 0 {
		return releasePath(p.owner, p.repo, p.releaseID)
	}
	if p.tag != "" {
		return releasePathByTag(p.owner, p.repo, p.tag)
	}
	if len(p.content) > 0 {
		return contentPath(p.owner, p.repo, p.branch, p.content)
	}
	if p.branch != "" {
		return branchPath(p.owner, p.repo, p.branch)
	}

	return repoPath(p.owner, p.repo)
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
				if id, err := strconv.ParseInt(parts[4], 10, 64); err == nil {
					p.releaseID = id
				}
			}
			if len(parts) > 5 {
				if id, err := strconv.ParseInt(parts[5], 10, 64); err == nil {
					p.assetID = id
				} else {
					p.asset = parts[5]
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
							if id, err := strconv.ParseInt(parts[5], 10, 64); err == nil {
								p.releaseID = id
							}
						}
					case "assets":
						if len(parts) > 5 {
							if id, err := strconv.ParseInt(parts[5], 10, 64); err == nil {
								p.assetID = id
							}
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

func contentPath(owner, repo, branch string, content []string) string {
	return fmt.Sprintf(
		"repos/%v/%v/contents/%v?ref=%v",
		owner, repo, strings.Join(content, "/"), branch,
	)
}

func releasePath(owner, repo string, id int64) string {
	return fmt.Sprintf(
		"repos/%v/%v/releases/%v",
		owner, repo, id,
	)
}

func releasePathByTag(owner, repo, tag string) string {
	return fmt.Sprintf(
		"repos/%v/%v/releases/tags/%v",
		owner, repo, url.PathEscape(tag),
	)
}

func assetPath(owner, repo string, id int64) string {
	return fmt.Sprintf("repos/%v/%v/releases/assets/%v", owner, repo, id)
}

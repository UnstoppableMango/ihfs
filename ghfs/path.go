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

func splitHost(u *url.URL) (host, rest string) {
	if host = u.Hostname(); host != "" {
		return host, u.RequestURI()
	}

	p := u.RequestURI()
	for _, h := range hosts {
		if after, ok := strings.CutPrefix(p, h); ok {
			return h, strings.TrimPrefix(after, "/")
		}
	}

	return "", p
}

// normalize returns the API path for name
func normalize(name string) (string, error) {
	u, err := url.Parse(name)
	if err != nil {
		return "", err
	}

	switch h, p := splitHost(u); h {
	case "api.github.com":
		return p, nil
	case "github.com":
		return fromWebURL(p)
	case "raw.githubusercontent.com":
		return fromRawURL(p)
	case "":
		return u.RequestURI(), nil
	default:
		return "", fmt.Errorf("invalid host: %s", h)
	}
}

func fromWebURL(name string) (string, error) {
	parts := strings.Split(clean(name), "/")
	for i, p := range parts {
		if unescaped, err := url.PathUnescape(p); err == nil {
			parts[i] = unescaped
		}
	}

	switch len(parts) {
	case 1:
		if p := parts[0]; p != "" {
			return ownerPath(p), nil
		}
		return "user", nil
	case 2:
		return repoPath(parts[0], parts[1]), nil
	case 4:
		// Expected pattern: owner/repo/tree/branch
		if parts[2] == "tree" {
			return branchPath(parts[0], parts[1], parts[3]), nil
		}
	case 5:
		// Expected patterns:
		// - owner/repo/blob/branch/path
		// - owner/repo/tree/branch/path
		// - owner/repo/releases/(tag|download)/TAG
		switch parts[2] {
		case "blob", "tree":
			return contentPath(parts[0], parts[1], parts[3], parts[4]), nil
		case "releases":
			if parts[3] == "tag" || parts[3] == "download" {
				return releasePath(parts[0], parts[1], parts[4]), nil
			}
		}
	}

	if len(parts) >= 6 {
		// Expected patterns:
		// - owner/repo/releases/(tag|download)/TAG/asset
		// - owner/repo/(tree|blob)/branch/path/to/item
		if parts[2] == "releases" && (parts[3] == "tag" || parts[3] == "download") {
			return assetPath(parts[0], parts[1], parts[4], parts[5]), nil
		}

		if parts[2] == "tree" || parts[2] == "blob" {
			return contentPath(parts[0], parts[1], parts[3],
				strings.Join(parts[4:], "/"),
			), nil
		}
	}

	return "", &ihfs.PathError{
		Op:   "open",
		Path: name,
		Err:  ihfs.ErrNotExist,
	}
}

func fromRawURL(name string) (string, error) {
	parts := strings.Split(clean(name), "/")
	for i, p := range parts {
		if unescaped, err := url.PathUnescape(p); err == nil {
			parts[i] = unescaped
		}
	}

	switch len(parts) {
	case 1:
		if p := parts[0]; p != "" {
			return ownerPath(p), nil
		}
		return "user", nil
	case 2:
		return repoPath(parts[0], parts[1]), nil
	case 3:
		return branchPath(parts[0], parts[1], parts[2]), nil
	}

	return contentPath(parts[0], parts[1], parts[2], strings.Join(parts[3:], "/")), nil
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

func ownerPath(owner string) string {
	return fmt.Sprintf("users/%v", owner)
}

func repoPath(owner, repository string) string {
	return fmt.Sprintf("repos/%v/%v", owner, repository)
}

func branchPath(owner, repository, branch string) string {
	return fmt.Sprintf("repos/%v/%v/branches/%v", owner, repository, url.PathEscape(branch))
}

func contentPath(owner, repository, branch, name string) string {
	segments := strings.Split(name, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return fmt.Sprintf(
		"repos/%v/%v/contents/%v?ref=%v",
		owner, repository, strings.Join(segments, "/"), url.QueryEscape(branch),
	)
}

func releasePath(owner, repository, name string) string {
	return fmt.Sprintf(
		"repos/%v/%v/releases/tags/%v",
		owner, repository, url.PathEscape(name),
	)
}

func assetPath(owner, repository, tag, name string) string {
	return fmt.Sprintf(
		"%v%v/%v/%v/%v",
		assetLookupPrefix,
		owner,
		repository,
		url.PathEscape(tag),
		url.PathEscape(name),
	)
}

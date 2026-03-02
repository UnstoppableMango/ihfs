package ghfs

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/unstoppablemango/ihfs"
)

var hosts = []string{
	"github.com",
	"api.github.com",
	"raw.githubusercontent.com",
}

func normalize(name string) (string, error) {
	if u, err := url.Parse(name); err == nil {
		switch u.Hostname() {
		case "api.github.com":
			return u.RequestURI(), nil
		case "github.com":
			return fromWebURL(u.Path)
		case "raw.githubusercontent.com":
			return fromRawURL(u.Path)
		case "":
			path := u.Path
			switch {
			case strings.HasPrefix(path, "github.com"):
				return fromWebURL(path)
			case strings.HasPrefix(path, "api.github.com"):
				cleaned := clean(path)
				if u.RawQuery != "" {
					return cleaned + "?" + u.RawQuery, nil
				}
				return cleaned, nil
			case strings.HasPrefix(path, "raw.githubusercontent.com"):
				return fromRawURL(path)
			default:
				return u.RequestURI(), nil
			}
		}
	}

	return name, nil
}

func fromWebURL(name string) (string, error) {
	parts := strings.Split(clean(name), "/")

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
		// - owner/repo/releases/(tag|download)/TAG
		if parts[2] == "blob" {
			return contentPath(parts[0], parts[1], parts[3], parts[4]), nil
		}
		return releasePath(parts[0], parts[1], parts[4]), nil
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
		path = u.Path
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
	return fmt.Sprintf("repos/%v/%v/branches/%v", owner, repository, branch)
}

func contentPath(owner, repository, branch, name string) string {
	return fmt.Sprintf(
		"repos/%v/%v/contents/%v?ref=%v",
		owner, repository, name, url.QueryEscape(branch),
	)
}

func releasePath(owner, repository, name string) string {
	return fmt.Sprintf(
		"repos/%v/%v/releases/tags/%v",
		owner, repository, name,
	)
}

func assetPath(owner, repository string, _ string, name string) string {
	return fmt.Sprintf(
		"repos/%v/%v/releases/assets/%v",
		owner, repository, name,
	)
}

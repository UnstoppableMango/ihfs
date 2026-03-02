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

func normalize(name string) (string, error) {
	u, err := url.Parse(name)
	if err != nil {
		return "", &ihfs.PathError{
			Op:   "open",
			Path: name,
			Err:  ihfs.ErrNotExist,
		}
	}

	switch u.Hostname() {
	case "api.github.com":
		return u.RequestURI(), nil
	case "github.com":
		return fromWebURL(u.EscapedPath())
	case "raw.githubusercontent.com":
		return fromRawURL(u.EscapedPath())
	case "":
		path := u.EscapedPath()
		switch {
		case path == "github.com" || strings.HasPrefix(path, "github.com/"):
			return fromWebURL(strings.TrimPrefix(path, "github.com"))
		case path == "api.github.com" || strings.HasPrefix(path, "api.github.com/"):
			cleaned := clean(strings.TrimPrefix(path, "api.github.com"))
			if u.RawQuery != "" {
				return cleaned + "?" + u.RawQuery, nil
			}
			return cleaned, nil
		case path == "raw.githubusercontent.com" || strings.HasPrefix(path, "raw.githubusercontent.com/"):
			return fromRawURL(strings.TrimPrefix(path, "raw.githubusercontent.com"))
		default:
			return u.RequestURI(), nil
		}
	default:
		return "", &ihfs.PathError{
			Op:   "open",
			Path: name,
			Err:  ihfs.ErrNotExist,
		}
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

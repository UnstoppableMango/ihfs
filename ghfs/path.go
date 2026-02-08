package ghfs

import (
	"net/url"
	"strings"
)

var hosts = []string{
	"github.com",
	"api.github.com",
	"raw.githubusercontent.com",
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

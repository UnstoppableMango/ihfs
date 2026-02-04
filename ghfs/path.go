package ghfs

import (
	"fmt"
	"strings"
)

type Path interface {
	fmt.Stringer

	Asset() *string
	Branch() *string
	Content() []string
	Owner() *string
	Repository() *string
	Release() *string
}

type ghpath struct {
	asset      *string
	branch     *string
	content    []string
	owner      *string
	repository *string
	release    *string
}

func parse(path string) (*ghpath, error) {
	var p ghpath
	var prev string

	for i, seg := range strings.Split(path, "/") {
		switch i {
		case 0:
			p.owner = &seg
		case 1:
			p.repository = &seg
		case 2:
			switch seg {
			case "releases", "tree", "refs":
				prev = seg
				continue
			default:
				return nil, fmt.Errorf("expected one of 'releases', 'tree', or 'refs', got: %s", seg)
			}
		case 3:
			// TODO: non-releases
			if seg == "tag" || seg == "download" {
				continue
			}
			return nil, fmt.Errorf("expected 'tag' or 'download' segment, got: %s", seg)
		case 4:
			// TODO: non-releases
			p.release = &seg
		case 5:
			// TODO: /download links
			p.asset = &seg
		}
	}

	return &p, nil
}

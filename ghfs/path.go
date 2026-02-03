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

func parse(path string) *ghpath {
	var p ghpath
	parts := strings.SplitN(path, "/", 4)
	if len(parts) > 0 {
		p.owner = &parts[0]
	}
	if len(parts) > 1 {
		p.repository = &parts[1]
	}
	if len(parts) > 2 {
		p.branch = &parts[2]
	}
	if len(parts) > 3 {
		p.content = strings.Split(parts[3], "/")
	}

	return &p
}

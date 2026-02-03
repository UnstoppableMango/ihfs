package ghfs

import "github.com/google/go-github/v82/github"

type File struct {
	name   string
	client *github.Client
}

package ghfs

import "fmt"

type Path interface {
	fmt.Stringer

	Asset() *string
	Branch() *string
	Content() []string
	Owner() *string
	Repository() *string
	Release() *string
}

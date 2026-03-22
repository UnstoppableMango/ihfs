package tarfs

import (
	"io/fs"
	"sync"
)

type entry interface {
	fs.File
	store(*sync.Map)
	within(dir string) bool
	dirEntry() fs.DirEntry
}

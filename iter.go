package ihfs

import (
	"io/fs"

	"github.com/unmango/go/iter"
)

type (
	// WalkFunc is an alias for fs.WalkDirFunc.
	WalkFunc = fs.WalkDirFunc
)

var (
	// SkipDir is an alias for fs.SkipDir.
	SkipDir = fs.SkipDir
)

// Iter returns a sequence that walks the file system fsys.
func Iter(fsys FS, root string) iter.Seq3[string, DirEntry, error] {
	return func(yield func(string, DirEntry, error) bool) {
		_ = Walk(fsys, root, func(path string, d fs.DirEntry, err error) error {
			if !yield(path, d, err) {
				return SkipDir
			}
			return nil
		})
	}
}

// IterPaths returns a sequence that walks the file system fsys, yielding paths.
func IterPaths(fsys FS, root string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		_ = Walk(fsys, root, func(path string, _ fs.DirEntry, err error) error {
			if !yield(path, err) {
				return SkipDir
			}
			return nil
		})
	}
}

// IterDirEntries returns a sequence that walks the file system fsys, yielding DirEntries.
func IterDirEntries(fsys FS, root string) iter.Seq2[DirEntry, error] {
	return func(yield func(DirEntry, error) bool) {
		_ = Walk(fsys, root, func(_ string, d fs.DirEntry, err error) error {
			if !yield(d, err) {
				return SkipDir
			}
			return nil
		})
	}
}

// Walk is a convenience wrapper around fs.WalkDir.
func Walk(fsys FS, root string, fn WalkFunc) error {
	return fs.WalkDir(fsys, root, fn)
}

package ihfs

import (
	"io/fs"

	"github.com/unmango/go/iter"
	"github.com/unmango/go/slices"
)

// WalkFunc is an alias for [fs.WalkDirFunc].
type WalkFunc = fs.WalkDirFunc

// SkipDir is an alias for [fs.SkipDir].
var SkipDir = fs.SkipDir

// Catch iterates over seq until an error occurs and returns the error and
// a Seq iterating over all paths and directory entries found before the error.
//
// When no error occurs, Catch returns a Seq that iterates over all paths and
// directory entries from seq.
func Catch(seq iter.Seq3[string, DirEntry, error]) (iter.Seq2[string, DirEntry], error) {
	var (
		final   error
		paths   []string
		entries []DirEntry
	)

	seq(func(path string, d DirEntry, err error) bool {
		final = err
		paths = append(paths, path)
		entries = append(entries, d)

		return err == nil
	})

	if final != nil {
		return nil, final
	} else {
		return slices.Zip(paths, entries), nil
	}
}

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

// Walk is a convenience wrapper around [fs.WalkDir].
func Walk(fsys FS, root string, fn WalkFunc) error {
	return fs.WalkDir(fsys, root, fn)
}

package try

import (
	"errors"
	"fmt"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/fsutil"
)

var ErrNotSupported = errors.New("operation not supported")

// DirExists reports if the given path exists and is a directory.
// If the FS does not implement [ihfs.Stat], DirExists returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func DirExists(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.Stat); ok {
		return fsutil.DirExists(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrNotSupported)
}

// Exists reports if the given path exists.
// If the FS does not implement [ihfs.Stat], Exists returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Exists(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.Stat); ok {
		return fsutil.Exists(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrNotSupported)
}

// IsDir reports if the given path exists and is a directory.
// If the FS does not implement [ihfs.Stat], IsDir returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func IsDir(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.Stat); ok {
		return fsutil.IsDir(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrNotSupported)
}

// ReadDir reads the named directory and returns a list of directory entries.
// If the FS does not implement [ihfs.ReadDir], ReadDir returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func ReadDir(fsys ihfs.FS, path string) ([]ihfs.DirEntry, error) {
	if dirfs, ok := fsys.(ihfs.ReadDir); ok {
		return dirfs.ReadDir(path)
	}
	return nil, fmt.Errorf("read dir: %w", ErrNotSupported)
}

// ReadDirNames reads the named directory and returns a list of names.
// If the FS does not implement [ihfs.ReadDir], ReadDirNames returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func ReadDirNames(fsys ihfs.FS, path string) ([]string, error) {
	if dirfs, ok := fsys.(ihfs.ReadDir); ok {
		return fsutil.ReadDirNames(dirfs, path)
	}
	return nil, fmt.Errorf("read dir: %w", ErrNotSupported)
}

// Stat attempts to call Stat on the given FS.
// If the FS does not implement [ihfs.Stat], Stat returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Stat(fsys ihfs.FS, path string) (ihfs.FileInfo, error) {
	if stat, ok := fsys.(ihfs.Stat); ok {
		return stat.Stat(path)
	}
	return nil, fmt.Errorf("stat: %w", ErrNotSupported)
}

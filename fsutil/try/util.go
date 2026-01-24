package try

import (
	"errors"
	"fmt"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/fsutil"
)

var ErrUnsupported = errors.New("operation not supported")

func DirExists(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.Stat); ok {
		return fsutil.DirExists(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrUnsupported)
}

func Exists(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.Stat); ok {
		return fsutil.Exists(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrUnsupported)
}

func IsDir(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.Stat); ok {
		return fsutil.IsDir(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrUnsupported)
}

func Stat(fsys ihfs.FS, path string) (ihfs.FileInfo, error) {
	if stat, ok := fsys.(ihfs.Stat); ok {
		return stat.Stat(path)
	}
	return nil, fmt.Errorf("stat: %w", ErrUnsupported)
}

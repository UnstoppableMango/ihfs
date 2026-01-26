package try

import (
	"errors"
	"fmt"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/fsutil"
)

var ErrNotSupported = errors.New("operation not supported")

func DirExists(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.Stat); ok {
		return fsutil.DirExists(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrNotSupported)
}

func Exists(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.Stat); ok {
		return fsutil.Exists(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrNotSupported)
}

func IsDir(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.Stat); ok {
		return fsutil.IsDir(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrNotSupported)
}

func ReadDir(fsys ihfs.FS, path string) ([]ihfs.DirEntry, error) {
	if dirfs, ok := fsys.(ihfs.ReadDir); ok {
		return dirfs.ReadDir(path)
	}
	return nil, fmt.Errorf("read dir: %w", ErrNotSupported)
}

func ReadDirNames(fsys ihfs.FS, path string) ([]string, error) {
	if dirfs, ok := fsys.(ihfs.ReadDir); ok {
		return fsutil.ReadDirNames(dirfs, path)
	}
	return nil, fmt.Errorf("read dir: %w", ErrNotSupported)
}

func Stat(fsys ihfs.FS, path string) (ihfs.FileInfo, error) {
	if stat, ok := fsys.(ihfs.Stat); ok {
		return stat.Stat(path)
	}
	return nil, fmt.Errorf("stat: %w", ErrNotSupported)
}

package fsutil

import (
	"errors"
	"io/fs"

	"github.com/unstoppablemango/ihfs"
)

func DirExists(fsys ihfs.Stat, path string) (bool, error) {
	info, err := fsys.Stat(path)
	if err == nil {
		return info.IsDir(), nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func Exists(fsys ihfs.Stat, path string) (bool, error) {
	_, err := fsys.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func IsDir(fsys ihfs.Stat, path string) (bool, error) {
	if info, err := fsys.Stat(path); err != nil {
		return false, err
	} else {
		return info.IsDir(), nil
	}
}

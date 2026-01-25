package fsutil

import (
	"errors"

	"github.com/unstoppablemango/ihfs"
)

func DirExists(fsys ihfs.Stat, path string) (bool, error) {
	info, err := fsys.Stat(path)
	if err == nil {
		return info.IsDir(), nil
	}
	if errors.Is(err, ihfs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func Exists(fsys ihfs.Stat, path string) (bool, error) {
	_, err := fsys.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ihfs.ErrNotExist) {
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

// ReadDirNames reads the named directory and returns a list of names.
func ReadDirNames(f ihfs.ReadDir, name string) ([]string, error) {
	entries, err := f.ReadDir(name)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names, nil
}

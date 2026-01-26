package fsutil

import (
	"errors"
	"fmt"
	"io"

	"github.com/unstoppablemango/ihfs"
)

// DirExists reports if the given path exists and is a directory.
//
// It differs from IsDir in that it returns false if the
// path does not exist, rather than returning an error.
func DirExists(fsys ihfs.Stat, path string) (bool, error) {
	if isDir, err := IsDir(fsys, path); err == nil {
		return isDir, nil
	} else if errors.Is(err, ihfs.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

// Exists reports if the given path exists.
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

// IsDir reports if the given path exists and is a directory.
// It calls fsys.Stat(path) and returns the result of FileInfo.IsDir().
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

// WriteReader reads all data from r and writes it to name in fsys using WriteFile.
// It returns an error if reading from r fails or if fsys.WriteFile reports an error.
func WriteReader(fsys ihfs.WriteFile, name string, r io.Reader, perm ihfs.FileMode) error {
	if data, err := io.ReadAll(r); err != nil {
		return fmt.Errorf("reading: %w", err)
	} else {
		return fsys.WriteFile(name, data, perm)
	}
}

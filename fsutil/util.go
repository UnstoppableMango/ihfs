package fsutil

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

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

// WriteReader reads all data from r and writes it to name in fsys using WriteFile.
// It returns an error if reading from r fails or if fsys.WriteFile reports an error.
func WriteReader(fsys ihfs.WriteFile, name string, r io.Reader) error {
	if data, err := io.ReadAll(r); err != nil {
		return fmt.Errorf("reading: %w", err)
	} else {
		return fsys.WriteFile(name, data, os.ModePerm)
	}
}

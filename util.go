package ihfs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
)

var ErrNotImplemented = errors.New("not implemented")

var (
	Glob    = fs.Glob
	ReadDir = fs.ReadDir
	Stat    = fs.Stat
)

// DirExists reports if the given path exists and is a directory.
//
// It differs from IsDir in that it returns false if the
// path does not exist, rather than returning an error.
func DirExists(fsys FS, path string) (bool, error) {
	if isDir, err := IsDir(fsys, path); err == nil {
		return isDir, nil
	} else if errors.Is(err, ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

// Exists reports if the given path exists.
func Exists(fsys FS, path string) (bool, error) {
	if _, err := Stat(fsys, path); err == nil {
		return true, nil
	} else if errors.Is(err, ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

// IsDir reports if the given path exists and is a directory.
// It calls [Stat] on fsys and returns the result of FileInfo.IsDir().
func IsDir(fsys FS, path string) (bool, error) {
	if info, err := Stat(fsys, path); err != nil {
		return false, err
	} else {
		return info.IsDir(), nil
	}
}

// Mkdir creates a new directory with the specified name and permission bits.
//
// If fsys implements [MkdirFS], Mkdir calls fsys.Mkdir.
// Otherwise, Mkdir returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func Mkdir(fsys FS, path string, perm FileMode) error {
	if fs, ok := fsys.(MkdirFS); ok {
		return fs.Mkdir(path, perm)
	}
	return fmt.Errorf("mkdir: %w", ErrNotImplemented)
}

// MkdirAll creates a new directory named path, along with any necessary parents, and sets permission bits.
//
// If fsys implements [MkdirAllFS], MkdirAll calls fsys.MkdirAll.
// Otherwise, MkdirAll attempts to create the directory and parents recursively.
func MkdirAll(fsys FS, path string, perm FileMode) error {
	if fs, ok := fsys.(MkdirAllFS); ok {
		return fs.MkdirAll(path, perm)
	}

	if path == "" {
		return nil
	}

	if err := Mkdir(fsys, path, perm); err != nil {
		if !errors.Is(err, ErrNotExist) {
			return err
		}

		parent := filepath.Dir(path)
		if parent == path {
			return err
		}
		if err := MkdirAll(fsys, parent, perm); err != nil {
			return err
		}
		return Mkdir(fsys, path, perm)
	}

	return nil
}

// ReadDirNames reads the named directory and returns a list of names.
func ReadDirNames(fsys FS, name string) ([]string, error) {
	entries, err := ReadDir(fsys, name)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names, nil
}

// WriteFile writes data to a file named by name in the given FS.
// The file mode perm is used when creating the file.
//
// If fsys implements [WriteFileFS], WriteFile calls fsys.WriteFile.
// Otherwise, WriteFile returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func WriteFile(fsys FS, name string, data []byte, perm FileMode) error {
	if wf, ok := fsys.(WriteFileFS); ok {
		return wf.WriteFile(name, data, perm)
	}
	return fmt.Errorf("write file: %w", ErrNotImplemented)
}

// WriteReader reads all data from r and writes it to name in fsys using [WriteFile].
// It returns an error if reading from r fails or if [WriteFile] reports an error.
func WriteReader(fsys FS, name string, r io.Reader, perm FileMode) error {
	if data, err := io.ReadAll(r); err != nil {
		return fmt.Errorf("reading: %w", err)
	} else {
		return WriteFile(fsys, name, data, perm)
	}
}

package ihfs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// ErrNotImplemented is returned when a filesystem operation is not supported.
var ErrNotImplemented = errors.New("not implemented")

// Convenience functions below use fallback strategies when the FS does not
// implement a specific interface. For example, Stat delegates to fs.Stat which
// may open the file and call Stat on the handle, and MkdirAll falls back to
// recursive Mkdir calls. For strict alternatives that return ErrNotImplemented
// instead of falling back, see the try package.

var (
	// Glob is an alias for [fs.Glob].
	Glob = fs.Glob
	// ReadDir is an alias for [fs.ReadDir].
	ReadDir = fs.ReadDir
	// Stat is an alias for [fs.Stat].
	Stat = fs.Stat
)

func Copy(dir string, fsys FS) error {
	if copier, ok := fsys.(CopyFS); ok {
		return copier.Copy(dir, fsys)
	}

	return Walk(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		destPath := filepath.Join(dir, path)

		if d.IsDir() {
			return os.MkdirAll(destPath, d.Type().Perm())
		}

		src, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		if _, err := os.Stat(destPath); err == nil {
			return &fs.PathError{Op: "copy", Path: destPath, Err: fs.ErrExist}
		}

		dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, d.Type().Perm())
		if err != nil {
			return err
		}
		defer dest.Close()

		if _, err := io.Copy(dest, src); err != nil {
			os.Remove(destPath)
			return err
		}

		return nil
	})
}

// DirExists reports if the given path exists and is a directory.
//
// It differs from IsDir in that it returns false if the
// path does not exist, rather than returning an error.
func DirExists(fsys FS, path string) (bool, error) {
	isDir, err := IsDir(fsys, path)
	if err == nil {
		return isDir, nil
	}
	if errors.Is(err, ErrNotExist) {
		return false, nil
	}
	return false, err
}

// Exists reports if the given path exists.
func Exists(fsys FS, path string) (bool, error) {
	_, err := Stat(fsys, path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrNotExist) {
		return false, nil
	}
	return false, err
}

// IsDir reports if the given path exists and is a directory.
// It calls [Stat] on fsys and returns the result of FileInfo.IsDir().
func IsDir(fsys FS, path string) (bool, error) {
	info, err := Stat(fsys, path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
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
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading: %w", err)
	}
	return WriteFile(fsys, name, data, perm)
}

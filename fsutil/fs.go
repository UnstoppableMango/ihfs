package fsutil

import (
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/unstoppablemango/ihfs"
)

// DirExists reports if the given path exists and is a directory.
//
// It differs from IsDir in that it returns false if the
// path does not exist, rather than returning an error.
func DirExists(fsys ihfs.StatFS, path string) (bool, error) {
	if isDir, err := IsDir(fsys, path); err == nil {
		return isDir, nil
	} else if errors.Is(err, ihfs.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

// Exists reports if the given path exists.
func Exists(fsys ihfs.StatFS, path string) (bool, error) {
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
func IsDir(fsys ihfs.StatFS, path string) (bool, error) {
	if info, err := fsys.Stat(path); err != nil {
		return false, err
	} else {
		return info.IsDir(), nil
	}
}

// Glob returns the names of all files matching pattern or nil
// if there is no matching file. The syntax of patterns is the same
// as in [path.Match]. The pattern may describe hierarchical names such as
// usr/*/bin/ed.
//
// Glob ignores file system errors such as I/O errors reading directories.
// The only possible returned error is [path.ErrBadPattern], reporting that
// the pattern is malformed.
//
// If fs implements [ihfs.GlobFS], Glob calls fs.Glob.
// Otherwise, Glob uses [ReadDir] to traverse the directory tree
// and look for matches for the pattern.
func Glob(fsys ihfs.ReadDirFS, pattern string) ([]string, error) {
	return fs.Glob(fsys, pattern)
}

// ReadDir reads the named directory
// and returns a list of directory entries sorted by filename.
//
// If fs implements [ihfs.ReadDirFS], ReadDir calls fs.ReadDir.
// Otherwise ReadDir calls fs.Open and uses ReadDir and Close
// on the returned file.
func ReadDir(fsys ihfs.FS, name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(fsys, name)
}

// ReadDirNames reads the named directory and returns a list of names.
func ReadDirNames(f ihfs.ReadDirFS, name string) ([]string, error) {
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

// Stat returns a [ihfs.FileInfo] describing the named file from the file system.
//
// If fs implements [ihfs.StatFS], Stat calls fs.Stat.
// Otherwise, Stat opens the [ihfs.File] to stat it.
func Stat(fsys ihfs.FS, name string) (ihfs.FileInfo, error) {
	return fs.Stat(fsys, name)
}

// WriteReader reads all data from r and writes it to name in fsys using WriteFile.
// It returns an error if reading from r fails or if fsys.WriteFile reports an error.
func WriteReader(fsys ihfs.WriteFileFS, name string, r io.Reader, perm ihfs.FileMode) error {
	if data, err := io.ReadAll(r); err != nil {
		return fmt.Errorf("reading: %w", err)
	} else {
		return fsys.WriteFile(name, data, perm)
	}
}

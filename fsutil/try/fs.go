package try

import (
	"errors"
	"fmt"
	"time"

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

// Chmod attempts to call Chmod on the given FS.
// If the FS does not implement [ihfs.Chmod], Chmod returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Chmod(fsys ihfs.FS, name string, mode ihfs.FileMode) error {
	if chmod, ok := fsys.(ihfs.Chmod); ok {
		return chmod.Chmod(name, mode)
	}
	return fmt.Errorf("chmod: %w", ErrNotSupported)
}

// Chown attempts to call Chown on the given FS.
// If the FS does not implement [ihfs.Chown], Chown returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Chown(fsys ihfs.FS, name string, uid, gid int) error {
	if chown, ok := fsys.(ihfs.Chown); ok {
		return chown.Chown(name, uid, gid)
	}
	return fmt.Errorf("chown: %w", ErrNotSupported)
}

// Chtimes attempts to call Chtimes on the given FS.
// If the FS does not implement [ihfs.Chtimes], Chtimes returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Chtimes(fsys ihfs.FS, name string, atime, mtime time.Time) error {
	if chtimes, ok := fsys.(ihfs.Chtimes); ok {
		return chtimes.Chtimes(name, atime, mtime)
	}
	return fmt.Errorf("chtimes: %w", ErrNotSupported)
}

// Copy attempts to call Copy on the given FS.
// If the FS does not implement [ihfs.Copy], Copy returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Copy(fsys ihfs.FS, dir string, src ihfs.FS) error {
	if copy, ok := fsys.(ihfs.Copy); ok {
		return copy.Copy(dir, src)
	}
	return fmt.Errorf("copy: %w", ErrNotSupported)
}

// Mkdir attempts to call Mkdir on the given FS.
// If the FS does not implement [ihfs.Mkdir], Mkdir returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Mkdir(fsys ihfs.FS, name string, mode ihfs.FileMode) error {
	if mkdir, ok := fsys.(ihfs.Mkdir); ok {
		return mkdir.Mkdir(name, mode)
	}
	return fmt.Errorf("mkdir: %w", ErrNotSupported)
}

// MkdirAll attempts to call MkdirAll on the given FS.
// If the FS does not implement [ihfs.MkdirAll], MkdirAll returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func MkdirAll(fsys ihfs.FS, name string, mode ihfs.FileMode) error {
	if mkdirAll, ok := fsys.(ihfs.MkdirAll); ok {
		return mkdirAll.MkdirAll(name, mode)
	}
	return fmt.Errorf("mkdir all: %w", ErrNotSupported)
}

// MkdirTemp attempts to call MkdirTemp on the given FS.
// If the FS does not implement [ihfs.MkdirTemp], MkdirTemp returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func MkdirTemp(fsys ihfs.FS, dir, pattern string) (string, error) {
	if mkdirTemp, ok := fsys.(ihfs.MkdirTemp); ok {
		return mkdirTemp.MkdirTemp(dir, pattern)
	}
	return "", fmt.Errorf("mkdir temp: %w", ErrNotSupported)
}

// Remove attempts to call Remove on the given FS.
// If the FS does not implement [ihfs.Remove], Remove returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Remove(fsys ihfs.FS, name string) error {
	if remove, ok := fsys.(ihfs.Remove); ok {
		return remove.Remove(name)
	}
	return fmt.Errorf("remove: %w", ErrNotSupported)
}

// RemoveAll attempts to call RemoveAll on the given FS.
// If the FS does not implement [ihfs.RemoveAll], RemoveAll returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func RemoveAll(fsys ihfs.FS, name string) error {
	if removeAll, ok := fsys.(ihfs.RemoveAll); ok {
		return removeAll.RemoveAll(name)
	}
	return fmt.Errorf("remove all: %w", ErrNotSupported)
}

// WriteFile attempts to call WriteFile on the given FS.
// If the FS does not implement [ihfs.WriteFile], WriteFile returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func WriteFile(fsys ihfs.FS, name string, data []byte, perm ihfs.FileMode) error {
	if writeFile, ok := fsys.(ihfs.WriteFile); ok {
		return writeFile.WriteFile(name, data, perm)
	}
	return fmt.Errorf("write file: %w", ErrNotSupported)
}

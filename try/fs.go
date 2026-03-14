package try

import (
	"fmt"
	"time"

	"github.com/unstoppablemango/ihfs"
)


var (
	// ErrNotSupported is deprecated: use ErrNotImplemented instead.
	ErrNotSupported = ihfs.ErrNotImplemented
	// ErrNotImplemented is returned when the filesystem does not implement the requested operation.
	ErrNotImplemented = ihfs.ErrNotImplemented
)

// Chmod attempts to call Chmod on the given FS.
// If the FS does not implement [ihfs.ChmodFS], Chmod returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Chmod(fsys ihfs.FS, name string, mode ihfs.FileMode) error {
	return ihfs.Chmod(fsys, name, mode)
}

// Chown attempts to call Chown on the given FS.
// If the FS does not implement [ihfs.ChownFS], Chown returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Chown(fsys ihfs.FS, name string, uid, gid int) error {
	return ihfs.Chown(fsys, name, uid, gid)
}

// Chtimes attempts to call Chtimes on the given FS.
// If the FS does not implement [ihfs.ChtimesFS], Chtimes returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Chtimes(fsys ihfs.FS, name string, atime, mtime time.Time) error {
	return ihfs.Chtimes(fsys, name, atime, mtime)
}

// Copy attempts to call Copy on the given FS.
// If the FS does not implement [ihfs.CopyFS], Copy returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Copy(fsys ihfs.FS, dir string, src ihfs.FS) error {
	if cfs, ok := fsys.(ihfs.CopyFS); ok {
		return cfs.Copy(dir, src)
	}
	return fmt.Errorf("copy: %w", ErrNotImplemented)
}

// Create attempts to call Create on the given FS.
// If the FS does not implement [ihfs.CreateFS], Create returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Create(fsys ihfs.FS, name string) (ihfs.File, error) {
	return ihfs.Create(fsys, name)
}

// CreateTemp attempts to call CreateTemp on the given FS.
// If the FS does not implement [ihfs.CreateTempFS], CreateTemp returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func CreateTemp(fsys ihfs.FS, dir, pattern string) (ihfs.File, error) {
	return ihfs.CreateTemp(fsys, dir, pattern)
}

// DirExists reports if the given path exists and is a directory.
// If the FS does not implement [ihfs.StatFS], DirExists returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func DirExists(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.StatFS); ok {
		return ihfs.DirExists(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrNotImplemented)
}

// Exists reports if the given path exists.
// If the FS does not implement [ihfs.StatFS], Exists returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Exists(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.StatFS); ok {
		return ihfs.Exists(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrNotImplemented)
}

// Glob attempts to call Glob on the given FS.
// If the FS does not implement [ihfs.GlobFS], Glob returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Glob(fsys ihfs.FS, pattern string) ([]string, error) {
	if glob, ok := fsys.(ihfs.GlobFS); ok {
		return glob.Glob(pattern)
	}
	return nil, fmt.Errorf("glob: %w", ErrNotImplemented)
}

// IsDir reports if the given path exists and is a directory.
// If the FS does not implement [ihfs.StatFS], IsDir returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func IsDir(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.StatFS); ok {
		return ihfs.IsDir(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrNotImplemented)
}

// Mkdir attempts to call Mkdir on the given FS.
// If the FS does not implement [ihfs.MkdirFS], Mkdir returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Mkdir(fsys ihfs.FS, name string, mode ihfs.FileMode) error {
	return ihfs.Mkdir(fsys, name, mode)
}

// MkdirAll attempts to call MkdirAll on the given FS.
// If the FS does not implement [ihfs.MkdirAllFS], MkdirAll returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func MkdirAll(fsys ihfs.FS, name string, mode ihfs.FileMode) error {
	if mkdirAll, ok := fsys.(ihfs.MkdirAllFS); ok {
		return mkdirAll.MkdirAll(name, mode)
	}
	return fmt.Errorf("mkdir all: %w", ErrNotImplemented)
}

// MkdirTemp attempts to call MkdirTemp on the given FS.
// If the FS does not implement [ihfs.MkdirTempFS], MkdirTemp returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func MkdirTemp(fsys ihfs.FS, dir, pattern string) (string, error) {
	return ihfs.MkdirTemp(fsys, dir, pattern)
}

// OpenFile attempts to call OpenFile on the given FS.
// If the FS does not implement [ihfs.OpenFileFS], OpenFile returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func OpenFile(fsys ihfs.FS, name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
	return ihfs.OpenFile(fsys, name, flag, perm)
}

// ReadDir reads the named directory and returns a list of directory entries.
// If the FS does not implement [ihfs.ReadDirFS], ReadDir returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func ReadDir(fsys ihfs.FS, path string) ([]ihfs.DirEntry, error) {
	if dirfs, ok := fsys.(ihfs.ReadDirFS); ok {
		return dirfs.ReadDir(path)
	}
	return nil, fmt.Errorf("read dir: %w", ErrNotImplemented)
}

// ReadDirNames reads the named directory and returns a list of names.
// If the FS does not implement [ihfs.ReadDirFS] or [ihfs.ReadDirNamesFS], ReadDirNames
// returns an error that can be checked with [errors.Is] for [ErrNotImplemented].
func ReadDirNames(fsys ihfs.FS, path string) ([]string, error) {
	if dirfs, ok := fsys.(ihfs.ReadDirNamesFS); ok {
		return dirfs.ReadDirNames(path)
	}
	if dirfs, ok := fsys.(ihfs.ReadDirFS); ok {
		return ihfs.ReadDirNames(dirfs, path)
	}
	return nil, fmt.Errorf("read dir: %w", ErrNotImplemented)
}

// ReadFile attempts to call ReadFile on the given FS.
// If the FS does not implement [ihfs.ReadFileFS], ReadFile returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func ReadFile(fsys ihfs.FS, name string) ([]byte, error) {
	return ihfs.ReadFile(fsys, name)
}

// ReadLink attempts to call ReadLink on the given FS.
// If the FS does not implement [ihfs.ReadLinkFS], ReadLink returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func ReadLink(fsys ihfs.FS, name string) (string, error) {
	return ihfs.ReadLink(fsys, name)
}

// Remove attempts to call Remove on the given FS.
// If the FS does not implement [ihfs.RemoveFS], Remove returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Remove(fsys ihfs.FS, name string) error {
	return ihfs.Remove(fsys, name)
}

// RemoveAll attempts to call RemoveAll on the given FS.
// If the FS does not implement [ihfs.RemoveAllFS], RemoveAll returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func RemoveAll(fsys ihfs.FS, name string) error {
	return ihfs.RemoveAll(fsys, name)
}

// Rename attempts to call Rename on the given FS.
// If the FS does not implement [ihfs.RenameFS], Rename returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Rename(fsys ihfs.FS, oldpath, newpath string) error {
	return ihfs.Rename(fsys, oldpath, newpath)
}

// Stat attempts to call Stat on the given FS.
// If the FS does not implement [ihfs.StatFS], Stat returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Stat(fsys ihfs.FS, name string) (ihfs.FileInfo, error) {
	if stat, ok := fsys.(ihfs.StatFS); ok {
		return stat.Stat(name)
	}
	return nil, fmt.Errorf("stat: %w", ErrNotImplemented)
}

// Sub attempts to call Sub on the given FS.
// If the FS does not implement [ihfs.SubFS], Sub returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Sub(fsys ihfs.FS, dir string) (ihfs.FS, error) {
	return ihfs.Sub(fsys, dir)
}

// Symlink attempts to call Symlink on the given FS.
// If the FS does not implement [ihfs.SymlinkFS], Symlink returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Symlink(fsys ihfs.FS, oldname, newname string) error {
	return ihfs.Symlink(fsys, oldname, newname)
}

// TempFile attempts to call TempFile on the given FS.
// If the FS does not implement [ihfs.TempFileFS], TempFile returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func TempFile(fsys ihfs.FS, dir, pattern string) (string, error) {
	return ihfs.TempFile(fsys, dir, pattern)
}

// WriteFile attempts to call WriteFile on the given FS.
// If the FS does not implement [ihfs.WriteFileFS], WriteFile returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func WriteFile(fsys ihfs.FS, name string, data []byte, perm ihfs.FileMode) error {
	return ihfs.WriteFile(fsys, name, data, perm)
}

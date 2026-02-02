package try

import (
	"fmt"
	"time"

	"github.com/unstoppablemango/ihfs"
)

var ErrNotImplemented = ihfs.ErrNotImplemented

// Chmod attempts to call Chmod on the given FS.
// If the FS does not implement [ihfs.ChmodFS], Chmod returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Chmod(fsys ihfs.FS, name string, mode ihfs.FileMode) error {
	if chmod, ok := fsys.(ihfs.ChmodFS); ok {
		return chmod.Chmod(name, mode)
	}
	return fmt.Errorf("chmod: %w", ErrNotImplemented)
}

// Chown attempts to call Chown on the given FS.
// If the FS does not implement [ihfs.ChownFS], Chown returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Chown(fsys ihfs.FS, name string, uid, gid int) error {
	if chown, ok := fsys.(ihfs.ChownFS); ok {
		return chown.Chown(name, uid, gid)
	}
	return fmt.Errorf("chown: %w", ErrNotImplemented)
}

// Chtimes attempts to call Chtimes on the given FS.
// If the FS does not implement [ihfs.ChtimesFS], Chtimes returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Chtimes(fsys ihfs.FS, name string, atime, mtime time.Time) error {
	if chtimes, ok := fsys.(ihfs.ChtimesFS); ok {
		return chtimes.Chtimes(name, atime, mtime)
	}
	return fmt.Errorf("chtimes: %w", ErrNotImplemented)
}

// Copy attempts to call Copy on the given FS.
// If the FS does not implement [ihfs.CopyFS], Copy returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Copy(fsys ihfs.FS, dir string, src ihfs.FS) error {
	if copy, ok := fsys.(ihfs.CopyFS); ok {
		return copy.Copy(dir, src)
	}
	return fmt.Errorf("copy: %w", ErrNotImplemented)
}

// Create attempts to call Create on the given FS.
// If the FS does not implement [ihfs.CreateFS], Create returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Create(fsys ihfs.FS, name string) (ihfs.File, error) {
	if create, ok := fsys.(ihfs.CreateFS); ok {
		return create.Create(name)
	}
	return nil, fmt.Errorf("create: %w", ErrNotImplemented)
}

// CreateTemp attempts to call CreateTemp on the given FS.
// If the FS does not implement [ihfs.CreateTempFS], CreateTemp returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func CreateTemp(fsys ihfs.FS, dir, pattern string) (ihfs.File, error) {
	if createTemp, ok := fsys.(ihfs.CreateTempFS); ok {
		return createTemp.CreateTemp(dir, pattern)
	}
	return nil, fmt.Errorf("create temp: %w", ErrNotImplemented)
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
	if mkdir, ok := fsys.(ihfs.MkdirFS); ok {
		return mkdir.Mkdir(name, mode)
	}
	return fmt.Errorf("mkdir: %w", ErrNotImplemented)
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
	if mkdirTemp, ok := fsys.(ihfs.MkdirTempFS); ok {
		return mkdirTemp.MkdirTemp(dir, pattern)
	}
	return "", fmt.Errorf("mkdir temp: %w", ErrNotImplemented)
}

// OpenFile attempts to call OpenFile on the given FS.
// If the FS does not implement [ihfs.OpenFileFS], OpenFile returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func OpenFile(fsys ihfs.FS, name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
	if openFile, ok := fsys.(ihfs.OpenFileFS); ok {
		return openFile.OpenFile(name, flag, perm)
	}
	return nil, fmt.Errorf("open file: %w", ErrNotImplemented)
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
	if readFile, ok := fsys.(ihfs.ReadFileFS); ok {
		return readFile.ReadFile(name)
	}
	return nil, fmt.Errorf("read file: %w", ErrNotImplemented)
}

// ReadLink attempts to call ReadLink on the given FS.
// If the FS does not implement [ihfs.ReadLinkFS], ReadLink returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func ReadLink(fsys ihfs.FS, name string) (string, error) {
	if readLink, ok := fsys.(ihfs.ReadLinkFS); ok {
		return readLink.ReadLink(name)
	}
	return "", fmt.Errorf("read link: %w", ErrNotImplemented)
}

// Remove attempts to call Remove on the given FS.
// If the FS does not implement [ihfs.RemoveFS], Remove returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Remove(fsys ihfs.FS, name string) error {
	if remove, ok := fsys.(ihfs.RemoveFS); ok {
		return remove.Remove(name)
	}
	return fmt.Errorf("remove: %w", ErrNotImplemented)
}

// RemoveAll attempts to call RemoveAll on the given FS.
// If the FS does not implement [ihfs.RemoveAllFS], RemoveAll returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func RemoveAll(fsys ihfs.FS, name string) error {
	if removeAll, ok := fsys.(ihfs.RemoveAllFS); ok {
		return removeAll.RemoveAll(name)
	}
	return fmt.Errorf("remove all: %w", ErrNotImplemented)
}

// Rename attempts to call Rename on the given FS.
// If the FS does not implement [ihfs.RenameFS], Rename returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Rename(fsys ihfs.FS, oldpath, newpath string) error {
	if rename, ok := fsys.(ihfs.RenameFS); ok {
		return rename.Rename(oldpath, newpath)
	}
	return fmt.Errorf("rename: %w", ErrNotImplemented)
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
	if sub, ok := fsys.(ihfs.SubFS); ok {
		return sub.Sub(dir)
	}
	return nil, fmt.Errorf("sub: %w", ErrNotImplemented)
}

// Symlink attempts to call Symlink on the given FS.
// If the FS does not implement [ihfs.SymlinkFS], Symlink returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Symlink(fsys ihfs.FS, oldname, newname string) error {
	if symlink, ok := fsys.(ihfs.SymlinkFS); ok {
		return symlink.Symlink(oldname, newname)
	}
	return fmt.Errorf("symlink: %w", ErrNotImplemented)
}

// TempFile attempts to call TempFile on the given FS.
// If the FS does not implement [ihfs.TempFileFS], TempFile returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func TempFile(fsys ihfs.FS, dir, pattern string) (string, error) {
	if tempFile, ok := fsys.(ihfs.TempFileFS); ok {
		return tempFile.TempFile(dir, pattern)
	}
	return "", fmt.Errorf("temp file: %w", ErrNotImplemented)
}

// WriteFile attempts to call WriteFile on the given FS.
// If the FS does not implement [ihfs.WriteFileFS], WriteFile returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func WriteFile(fsys ihfs.FS, name string, data []byte, perm ihfs.FileMode) error {
	if writeFile, ok := fsys.(ihfs.WriteFileFS); ok {
		return writeFile.WriteFile(name, data, perm)
	}
	return fmt.Errorf("write file: %w", ErrNotImplemented)
}

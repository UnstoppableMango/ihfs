package try

import (
	"errors"
	"fmt"
	"time"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/fsutil"
)

var ErrNotSupported = errors.New("operation not supported")

// Chmod attempts to call Chmod on the given FS.
// If the FS does not implement [ihfs.ChmodFS], Chmod returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Chmod(fsys ihfs.FS, name string, mode ihfs.FileMode) error {
	if chmod, ok := fsys.(ihfs.ChmodFS); ok {
		return chmod.Chmod(name, mode)
	}
	return fmt.Errorf("chmod: %w", ErrNotSupported)
}

// Chown attempts to call Chown on the given FS.
// If the FS does not implement [ihfs.ChownFS], Chown returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Chown(fsys ihfs.FS, name string, uid, gid int) error {
	if chown, ok := fsys.(ihfs.ChownFS); ok {
		return chown.Chown(name, uid, gid)
	}
	return fmt.Errorf("chown: %w", ErrNotSupported)
}

// Chtimes attempts to call Chtimes on the given FS.
// If the FS does not implement [ihfs.ChtimesFS], Chtimes returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Chtimes(fsys ihfs.FS, name string, atime, mtime time.Time) error {
	if chtimes, ok := fsys.(ihfs.ChtimesFS); ok {
		return chtimes.Chtimes(name, atime, mtime)
	}
	return fmt.Errorf("chtimes: %w", ErrNotSupported)
}

// Copy attempts to call Copy on the given FS.
// If the FS does not implement [ihfs.CopyFS], Copy returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Copy(fsys ihfs.FS, dir string, src ihfs.FS) error {
	if copy, ok := fsys.(ihfs.CopyFS); ok {
		return copy.Copy(dir, src)
	}
	return fmt.Errorf("copy: %w", ErrNotSupported)
}

// Create attempts to call Create on the given FS.
// If the FS does not implement [ihfs.CreateFS], Create returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Create(fsys ihfs.FS, name string) (ihfs.File, error) {
	if create, ok := fsys.(ihfs.CreateFS); ok {
		return create.Create(name)
	}
	return nil, fmt.Errorf("create: %w", ErrNotSupported)
}

// CreateTemp attempts to call CreateTemp on the given FS.
// If the FS does not implement [ihfs.CreateTempFS], CreateTemp returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func CreateTemp(fsys ihfs.FS, dir, pattern string) (ihfs.File, error) {
	if createTemp, ok := fsys.(ihfs.CreateTempFS); ok {
		return createTemp.CreateTemp(dir, pattern)
	}
	return nil, fmt.Errorf("create temp: %w", ErrNotSupported)
}

// DirExists reports if the given path exists and is a directory.
// If the FS does not implement [ihfs.StatFS], DirExists returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func DirExists(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.StatFS); ok {
		return fsutil.DirExists(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrNotSupported)
}

// Exists reports if the given path exists.
// If the FS does not implement [ihfs.StatFS], Exists returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Exists(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.StatFS); ok {
		return fsutil.Exists(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrNotSupported)
}

// Glob attempts to call Glob on the given FS.
// If the FS does not implement [ihfs.GlobFS], Glob returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Glob(fsys ihfs.FS, pattern string) ([]string, error) {
	if glob, ok := fsys.(ihfs.GlobFS); ok {
		return glob.Glob(pattern)
	}
	return nil, fmt.Errorf("glob: %w", ErrNotSupported)
}

// IsDir reports if the given path exists and is a directory.
// If the FS does not implement [ihfs.StatFS], IsDir returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func IsDir(fsys ihfs.FS, path string) (bool, error) {
	if stat, ok := fsys.(ihfs.StatFS); ok {
		return fsutil.IsDir(stat, path)
	}
	return false, fmt.Errorf("stat: %w", ErrNotSupported)
}

// Mkdir attempts to call Mkdir on the given FS.
// If the FS does not implement [ihfs.MkdirFS], Mkdir returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Mkdir(fsys ihfs.FS, name string, mode ihfs.FileMode) error {
	if mkdir, ok := fsys.(ihfs.MkdirFS); ok {
		return mkdir.Mkdir(name, mode)
	}
	return fmt.Errorf("mkdir: %w", ErrNotSupported)
}

// MkdirAll attempts to call MkdirAll on the given FS.
// If the FS does not implement [ihfs.MkdirAllFS], MkdirAll returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func MkdirAll(fsys ihfs.FS, name string, mode ihfs.FileMode) error {
	if mkdirAll, ok := fsys.(ihfs.MkdirAllFS); ok {
		return mkdirAll.MkdirAll(name, mode)
	}
	return fmt.Errorf("mkdir all: %w", ErrNotSupported)
}

// MkdirTemp attempts to call MkdirTemp on the given FS.
// If the FS does not implement [ihfs.MkdirTempFS], MkdirTemp returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func MkdirTemp(fsys ihfs.FS, dir, pattern string) (string, error) {
	if mkdirTemp, ok := fsys.(ihfs.MkdirTempFS); ok {
		return mkdirTemp.MkdirTemp(dir, pattern)
	}
	return "", fmt.Errorf("mkdir temp: %w", ErrNotSupported)
}

// OpenFile attempts to call OpenFile on the given FS.
// If the FS does not implement [ihfs.OpenFileFS], OpenFile returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func OpenFile(fsys ihfs.FS, name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
	if openFile, ok := fsys.(ihfs.OpenFileFS); ok {
		return openFile.OpenFile(name, flag, perm)
	}
	return nil, fmt.Errorf("open file: %w", ErrNotSupported)
}

// ReadDir reads the named directory and returns a list of directory entries.
// If the FS does not implement [ihfs.ReadDirFS], ReadDir returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func ReadDir(fsys ihfs.FS, path string) ([]ihfs.DirEntry, error) {
	if dirfs, ok := fsys.(ihfs.ReadDirFS); ok {
		return dirfs.ReadDir(path)
	}
	return nil, fmt.Errorf("read dir: %w", ErrNotSupported)
}

// ReadDirNames reads the named directory and returns a list of names.
// If the FS does not implement [ihfs.ReadDirFS], ReadDirNames returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func ReadDirNames(fsys ihfs.FS, path string) ([]string, error) {
	if dirfs, ok := fsys.(ihfs.ReadDirNamesFS); ok {
		return dirfs.ReadDirNames(path)
	}
	if dirfs, ok := fsys.(ihfs.ReadDirFS); ok {
		return fsutil.ReadDirNames(dirfs, path)
	}
	return nil, fmt.Errorf("read dir: %w", ErrNotSupported)
}

// ReadFile attempts to call ReadFile on the given FS.
// If the FS does not implement [ihfs.ReadFileFS], ReadFile returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func ReadFile(fsys ihfs.FS, name string) ([]byte, error) {
	if readFile, ok := fsys.(ihfs.ReadFileFS); ok {
		return readFile.ReadFile(name)
	}
	return nil, fmt.Errorf("read file: %w", ErrNotSupported)
}

// ReadLink attempts to call ReadLink on the given FS.
// If the FS does not implement [ihfs.ReadLinkFS], ReadLink returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func ReadLink(fsys ihfs.FS, name string) (string, error) {
	if readLink, ok := fsys.(ihfs.ReadLinkFS); ok {
		return readLink.ReadLink(name)
	}
	return "", fmt.Errorf("read link: %w", ErrNotSupported)
}

// Remove attempts to call Remove on the given FS.
// If the FS does not implement [ihfs.RemoveFS], Remove returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Remove(fsys ihfs.FS, name string) error {
	if remove, ok := fsys.(ihfs.RemoveFS); ok {
		return remove.Remove(name)
	}
	return fmt.Errorf("remove: %w", ErrNotSupported)
}

// RemoveAll attempts to call RemoveAll on the given FS.
// If the FS does not implement [ihfs.RemoveAllFS], RemoveAll returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func RemoveAll(fsys ihfs.FS, name string) error {
	if removeAll, ok := fsys.(ihfs.RemoveAllFS); ok {
		return removeAll.RemoveAll(name)
	}
	return fmt.Errorf("remove all: %w", ErrNotSupported)
}

// Rename attempts to call Rename on the given FS.
// If the FS does not implement [ihfs.RenameFS], Rename returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Rename(fsys ihfs.FS, oldpath, newpath string) error {
	if rename, ok := fsys.(ihfs.RenameFS); ok {
		return rename.Rename(oldpath, newpath)
	}
	return fmt.Errorf("rename: %w", ErrNotSupported)
}

// Stat attempts to call Stat on the given FS.
// If the FS does not implement [ihfs.StatFS], Stat returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Stat(fsys ihfs.FS, path string) (ihfs.FileInfo, error) {
	if stat, ok := fsys.(ihfs.StatFS); ok {
		return stat.Stat(path)
	}
	return nil, fmt.Errorf("stat: %w", ErrNotSupported)
}

// Sub attempts to call Sub on the given FS.
// If the FS does not implement [ihfs.SubFS], Sub returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Sub(fsys ihfs.FS, dir string) (ihfs.FS, error) {
	if sub, ok := fsys.(ihfs.SubFS); ok {
		return sub.Sub(dir)
	}
	return nil, fmt.Errorf("sub: %w", ErrNotSupported)
}

// Symlink attempts to call Symlink on the given FS.
// If the FS does not implement [ihfs.SymlinkFS], Symlink returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Symlink(fsys ihfs.FS, oldname, newname string) error {
	if symlink, ok := fsys.(ihfs.SymlinkFS); ok {
		return symlink.Symlink(oldname, newname)
	}
	return fmt.Errorf("symlink: %w", ErrNotSupported)
}

// TempFile attempts to call TempFile on the given FS.
// If the FS does not implement [ihfs.TempFileFS], TempFile returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func TempFile(fsys ihfs.FS, dir, pattern string) (string, error) {
	if tempFile, ok := fsys.(ihfs.TempFileFS); ok {
		return tempFile.TempFile(dir, pattern)
	}
	return "", fmt.Errorf("temp file: %w", ErrNotSupported)
}

// WriteFile attempts to call WriteFile on the given FS.
// If the FS does not implement [ihfs.WriteFileFS], WriteFile returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func WriteFile(fsys ihfs.FS, name string, data []byte, perm ihfs.FileMode) error {
	if writeFile, ok := fsys.(ihfs.WriteFileFS); ok {
		return writeFile.WriteFile(name, data, perm)
	}
	return fmt.Errorf("write file: %w", ErrNotSupported)
}

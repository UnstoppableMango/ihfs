package ihfs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"
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

// Copy copies the contents of src into dest under the directory prefix dir.
// If dest implements [CopyFS], the operation is delegated to its Copy method.
// Otherwise Copy walks src and recreates directories, regular files, and
// symbolic links in dest. Existing files are not overwritten: regular files
// are created with os.O_CREATE|os.O_EXCL and the call fails if the target
// path already exists.
//
// Directories are created via MkdirAll with permissions derived from the
// source entry's mode (defaulting to 0755 if the mode has no permission bits).
// Regular files are created with permissions based on the source file's mode.
// Symbolic links are reproduced only if dest implements [SymlinkFS]; otherwise
// Copy returns an error wrapping ErrNotImplemented.
//
// Copy propagates errors from the underlying filesystem operations. For some
// failures it returns an [fs.PathError] identifying the operation and path
// that caused the error.
func Copy(dest FS, dir string, src FS) error {
	if copier, ok := dest.(CopyFS); ok {
		return copier.Copy(dir, src)
	}

	return Walk(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		destPath := path.Join(dir, p)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			perm := info.Mode().Perm()
			if perm == 0 {
				perm = 0755
			}
			return MkdirAll(dest, destPath, perm)
		}

		switch d.Type() {
		case fs.ModeSymlink:
			target, err := ReadLink(src, p)
			if err != nil {
				return fmt.Errorf("copy: %w", err)
			}
			if linker, ok := dest.(SymlinkFS); ok {
				return linker.Symlink(target, destPath)
			}
			return fmt.Errorf("copy: symlink: %w", ErrNotImplemented)
		case 0: // regular file
			r, err := src.Open(p)
			if err != nil {
				return err
			}
			defer func() { _ = r.Close() }()

			info, err := r.Stat()
			if err != nil {
				return err
			}

			perm := info.Mode().Perm()
			if perm == 0 {
				perm = 0644
			}

			w, err := OpenFile(dest, destPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
			if err != nil {
				return err
			}

			writer, ok := w.(io.Writer)
			if !ok {
				_ = w.Close()
				return fmt.Errorf("copy: %w", ErrNotImplemented)
			}

			if _, err := io.Copy(writer, r); err != nil {
				_ = w.Close()
				return &fs.PathError{Op: "Copy", Path: destPath, Err: err}
			}
			return w.Close()
		default:
			return &fs.PathError{Op: "Copy", Path: p, Err: fs.ErrInvalid}
		}
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

// OpenFile opens the named file with specified flag (i.e. O_RDONLY, O_WRONLY, O_RDWR) and permission (before umask).
//
// If the FS does not implement [OpenFileFS], OpenFile returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func OpenFile(fsys FS, name string, flag int, perm FileMode) (File, error) {
	if opener, ok := fsys.(OpenFileFS); ok {
		return opener.OpenFile(name, flag, perm)
	}
	return nil, fmt.Errorf("open file: %w", ErrNotImplemented)
}

// Remove removes the named file or (empty) directory.
//
// If the FS does not implement [RemoveFS], Remove returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Remove(fsys FS, name string) error {
	if remover, ok := fsys.(RemoveFS); ok {
		return remover.Remove(name)
	}
	return fmt.Errorf("remove: %w", ErrNotImplemented)
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

// Chmod changes the mode of the named file in fsys.
//
// If fsys implements [ChmodFS], Chmod calls fsys.Chmod.
// Otherwise, Chmod returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func Chmod(fsys FS, name string, mode FileMode) error {
	if chmod, ok := fsys.(ChmodFS); ok {
		return chmod.Chmod(name, mode)
	}
	return fmt.Errorf("chmod: %w", ErrNotImplemented)
}

// Chown changes the uid and gid of the named file in fsys.
//
// If fsys implements [ChownFS], Chown calls fsys.Chown.
// Otherwise, Chown returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func Chown(fsys FS, name string, uid, gid int) error {
	if chown, ok := fsys.(ChownFS); ok {
		return chown.Chown(name, uid, gid)
	}
	return fmt.Errorf("chown: %w", ErrNotImplemented)
}

// Chtimes changes the access and modification times of the named file in fsys.
//
// If fsys implements [ChtimesFS], Chtimes calls fsys.Chtimes.
// Otherwise, Chtimes returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func Chtimes(fsys FS, name string, atime, mtime time.Time) error {
	if chtimes, ok := fsys.(ChtimesFS); ok {
		return chtimes.Chtimes(name, atime, mtime)
	}
	return fmt.Errorf("chtimes: %w", ErrNotImplemented)
}

// Create creates the named file in fsys.
//
// If fsys implements [CreateFS], Create calls fsys.Create.
// Otherwise, Create returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func Create(fsys FS, name string) (File, error) {
	if create, ok := fsys.(CreateFS); ok {
		return create.Create(name)
	}
	return nil, fmt.Errorf("create: %w", ErrNotImplemented)
}

// CreateTemp creates a temporary file in fsys.
//
// If fsys implements [CreateTempFS], CreateTemp calls fsys.CreateTemp.
// Otherwise, CreateTemp returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func CreateTemp(fsys FS, dir, pattern string) (File, error) {
	if createTemp, ok := fsys.(CreateTempFS); ok {
		return createTemp.CreateTemp(dir, pattern)
	}
	return nil, fmt.Errorf("create temp: %w", ErrNotImplemented)
}

// MkdirTemp creates a temporary directory in fsys.
//
// If fsys implements [MkdirTempFS], MkdirTemp calls fsys.MkdirTemp.
// Otherwise, MkdirTemp returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func MkdirTemp(fsys FS, dir, pattern string) (string, error) {
	if mkdirTemp, ok := fsys.(MkdirTempFS); ok {
		return mkdirTemp.MkdirTemp(dir, pattern)
	}
	return "", fmt.Errorf("mkdir temp: %w", ErrNotImplemented)
}

// ReadFile reads the named file in fsys.
//
// If fsys implements [ReadFileFS], ReadFile calls fsys.ReadFile.
// Otherwise, ReadFile returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func ReadFile(fsys FS, name string) ([]byte, error) {
	if readFile, ok := fsys.(ReadFileFS); ok {
		return readFile.ReadFile(name)
	}
	return nil, fmt.Errorf("read file: %w", ErrNotImplemented)
}

// ReadLink reads the target of the symbolic link named by name in fsys.
//
// If fsys implements [ReadLinkFS], ReadLink calls fsys.ReadLink.
// Otherwise, ReadLink returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func ReadLink(fsys FS, name string) (string, error) {
	if readLink, ok := fsys.(ReadLinkFS); ok {
		return readLink.ReadLink(name)
	}
	return "", fmt.Errorf("read link: %w", ErrNotImplemented)
}

// RemoveAll removes the named path and all its children from fsys.
//
// If fsys implements [RemoveAllFS], RemoveAll calls fsys.RemoveAll.
// Otherwise, RemoveAll returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func RemoveAll(fsys FS, name string) error {
	if removeAll, ok := fsys.(RemoveAllFS); ok {
		return removeAll.RemoveAll(name)
	}
	return fmt.Errorf("remove all: %w", ErrNotImplemented)
}

// Rename renames (moves) oldpath to newpath in fsys.
//
// If fsys implements [RenameFS], Rename calls fsys.Rename.
// Otherwise, Rename returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func Rename(fsys FS, oldpath, newpath string) error {
	if rename, ok := fsys.(RenameFS); ok {
		return rename.Rename(oldpath, newpath)
	}
	return fmt.Errorf("rename: %w", ErrNotImplemented)
}

// Sub returns an FS rooted at fsys's dir subtree.
//
// If fsys implements [SubFS], Sub calls fsys.Sub.
// Otherwise, Sub returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func Sub(fsys FS, dir string) (FS, error) {
	if sub, ok := fsys.(SubFS); ok {
		return sub.Sub(dir)
	}
	return nil, fmt.Errorf("sub: %w", ErrNotImplemented)
}

// Symlink creates a symbolic link named newname pointing to oldname in fsys.
//
// If fsys implements [SymlinkFS], Symlink calls fsys.Symlink.
// Otherwise, Symlink returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func Symlink(fsys FS, oldname, newname string) error {
	if symlink, ok := fsys.(SymlinkFS); ok {
		return symlink.Symlink(oldname, newname)
	}
	return fmt.Errorf("symlink: %w", ErrNotImplemented)
}

// TempFile returns the path of a new temporary file in fsys.
//
// If fsys implements [TempFileFS], TempFile calls fsys.TempFile.
// Otherwise, TempFile returns an error that can be checked
// with [errors.Is] for [ErrNotImplemented].
func TempFile(fsys FS, dir, pattern string) (string, error) {
	if tempFile, ok := fsys.(TempFileFS); ok {
		return tempFile.TempFile(dir, pattern)
	}
	return "", fmt.Errorf("temp file: %w", ErrNotImplemented)
}

// WriteReader reads all data from r and writes it to name in fsys using [WriteFile].
// It returns an error if reading from r fails or if [WriteFile] reports an error.
func WriteReader(fsys FS, name string, r io.Reader, perm FileMode) error {
	// TODO: probably a more efficient way to do this
	if data, err := io.ReadAll(r); err != nil {
		return fmt.Errorf("reading: %w", err)
	} else {
		return WriteFile(fsys, name, data, perm)
	}
}

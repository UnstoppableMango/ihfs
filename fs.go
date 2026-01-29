package ihfs

import (
	"io/fs"
	"time"

	"github.com/unmango/go/os"
)

type (
	FS         = fs.FS
	GlobFS     = fs.GlobFS
	OsFS       = os.Fs
	ReadDirFS  = fs.ReadDirFS
	ReadFileFS = fs.ReadFileFS
	ReadLinkFS = fs.ReadLinkFS
	StatFS     = fs.StatFS
	SubFS      = fs.SubFS
)

// Ensure interface compliance with [os.Os].
var _ FS = (OsFS)(nil)

// ChmodFS is the interface implemented by a file system that supports changing file modes.
type ChmodFS interface {
	FS

	// Chmod changes the mode of the named file to mode.
	// If the file is a symbolic link, it changes the mode of the link's target.
	// If there is an error, it should be of type [*PathError].
	Chmod(name string, mode FileMode) error
}

// ChownFS is the interface implemented by a file system that supports changing file ownership.
type ChownFS interface {
	FS

	// Chown changes the numeric uid and gid of the named file.
	// If the file is a symbolic link, it changes the uid and gid of the link's target.
	// A uid or gid of -1 means to not change that value.
	// If there is an error, it should be of type [*PathError].
	Chown(name string, uid, gid int) error
}

// ChtimesFS is the interface implemented by a file system that supports changing file access and modification times.
type ChtimesFS interface {
	FS

	// Chtimes changes the access and modification times of the named
	// file, similar to the Unix utime() or utimes() functions.
	// A zero [time.Time] value should leave the corresponding file time unchanged.
	//
	// The underlying filesystem may truncate or round the values to a
	// less precise time unit.
	// If there is an error, it should be of type [*PathError].
	Chtimes(name string, atime, mtime time.Time) error
}

// CopyFS is the interface implemented by a file system that supports copying another file system.
type CopyFS interface {
	FS

	// Copy copies the file system fsys into the directory dir.
	// Implementations should create dir if necessary.
	//
	// Copy should not overwrite existing files. If a file name in fsys
	// already exists in the destination, Copy should return an error
	// such that errors.Is(err, fs.ErrExist) will be true.
	//
	// Symbolic links in dir should be followed.
	//
	// Copying should stop at and return the first error encountered.
	Copy(dir string, fsys FS) error
}

// Create is the interface implemented by a file system that supports creating new files.
type Create interface {
	FS

	// Create creates a new file with the specified name.
	// If the file already exists, it should be truncated.
	// If there is an error, it should be of type [*PathError].
	Create(name string) (File, error)
}

// Linker is the interface implemented by a file system that supports creating and reading symbolic links.
type Linker interface {
	Symlink
	ReadLink
}

// MkdirFS is the interface implemented by a file system that supports creating directories.
type MkdirFS interface {
	FS

	// Mkdir creates a new directory with the specified name and permission
	// bits (before umask).
	// If there is an error, it should be of type [*PathError].
	Mkdir(name string, mode FileMode) error
}

// MkdirAllFS is the interface implemented by a file system that supports creating directories along a path.
type MkdirAllFS interface {
	FS

	// MkdirAll creates a directory named path,
	// along with any necessary parents, and should return nil,
	// or else returns an error.
	// The permission bits perm (before umask) should be used for all
	// directories that MkdirAll creates.
	// If path is already a directory, MkdirAll should do nothing
	// and return nil.
	MkdirAll(name string, mode FileMode) error
}

// MkdirTempFS is the interface implemented by a file system that supports creating temporary directories.
type MkdirTempFS interface {
	FS

	// MkdirTemp creates a new temporary directory in the directory dir
	// and returns the pathname of the new directory.
	// It is the caller's responsibility to remove the directory when it is no longer needed.
	MkdirTemp(dir, pattern string) (name string, err error)
}

// OpenFile is the interface implemented by a file system that supports opening files.
type OpenFile interface {
	FS

	// OpenFile opens the named file with specified flag (O_RDONLY, O_WRONLY, O_RDWR) and permission (before umask).
	// If there is an error, it should be of type [*PathError].
	OpenFile(name string, flag int, perm FileMode) (File, error)
}

// RemoveFS is the interface implemented by a file system that supports removing files.
type RemoveFS interface {
	FS

	// Remove removes the named file or (empty) directory.
	// If there is an error, it should be of type [*PathError].
	Remove(name string) error
}

// RemoveAllFS is the interface implemented by a file system that supports removing directories and their contents.
type RemoveAllFS interface {
	FS

	// RemoveAll removes path and any children it contains.
	// It removes everything it can but returns the first error
	// it encounters. If the path does not exist, RemoveAll
	// should return nil (no error).
	// If there is an error, it should be of type [*PathError].
	RemoveAll(name string) error
}

// Rename is the interface implemented by a file system that supports renaming files.
type Rename interface {
	FS

	// Rename renames (moves) oldpath to newpath.
	// If there is an error, it should be of type [*PathError].
	Rename(oldpath, newpath string) error
}

// Symlink is the interface implemented by a file system that supports creating symbolic links.
type Symlink interface {
	FS

	// Symlink creates a symbolic link named newname pointing to oldname.
	// If there is an error, it should be of type [*PathError].
	Symlink(oldname, newname string) error
}

// TempFile is the interface implemented by a file system that supports creating temporary files.
type TempFile interface {
	FS

	// TempFile creates a new temporary file in the directory dir
	// and returns the pathname of the new file.
	// It is the caller's responsibility to remove the file when it is no longer needed.
	TempFile(dir, pattern string) (name string, err error)
}

// WriteFileFS is the interface implemented by a file system that supports writing files.
type WriteFileFS interface {
	FS

	// WriteFile writes data to the named file.
	WriteFile(name string, data []byte, perm FileMode) error
}

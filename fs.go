package ihfs

import (
	"io/fs"
	"time"

	"github.com/unmango/go/os"
)

type (
	FS       = fs.FS
	Glob     = fs.GlobFS
	Os       = os.Fs
	ReadDir  = fs.ReadDirFS
	ReadFile = fs.ReadFileFS
	ReadLink = fs.ReadLinkFS
	Stat     = fs.StatFS
	Sub      = fs.SubFS
)

var (
	ErrClosed     = fs.ErrClosed
	ErrExist      = fs.ErrExist
	ErrInvalid    = fs.ErrInvalid
	ErrNotExist   = fs.ErrNotExist
	ErrPermission = fs.ErrPermission
)

// Operation represents a file system operation.
type Operation interface {
	// Subject returns the subject of the operation, typically a file or directory path.
	Subject() string
}

// Ensure interface compliance with [os.Os].
var _ FS = (Os)(nil)

// Chmod is the interface implemented by a file system that supports changing file modes.
type Chmod interface {
	FS

	// Chmod changes the mode of the named file to mode.
	// If the file is a symbolic link, it changes the mode of the link's target.
	// If there is an error, it should be of type [*PathError].
	Chmod(name string, mode FileMode) error
}

// Chown is the interface implemented by a file system that supports changing file ownership.
type Chown interface {
	FS

	// Chown changes the numeric uid and gid of the named file.
	// If the file is a symbolic link, it changes the uid and gid of the link's target.
	// A uid or gid of -1 means to not change that value.
	// If there is an error, it should be of type [*PathError].
	Chown(name string, uid, gid int) error
}

// Chtimes is the interface implemented by a file system that supports changing file access and modification times.
type Chtimes interface {
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

// Copy is the interface implemented by a file system that supports copying another file system.
type Copy interface {
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

// Mkdir is the interface implemented by a file system that supports creating directories.
type Mkdir interface {
	FS

	// Mkdir creates a new directory with the specified name and permission
	// bits (before umask).
	// If there is an error, it should be of type [*PathError].
	Mkdir(name string, mode FileMode) error
}

// MkdirAll is the interface implemented by a file system that supports creating directories along a path.
type MkdirAll interface {
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

// MkdirTemp is the interface implemented by a file system that supports creating temporary directories.
type MkdirTemp interface {
	FS

	// MkdirTemp creates a new temporary directory in the directory dir
	// and returns the pathname of the new directory.
	// It is the caller's responsibility to remove the directory when it is no longer needed.
	MkdirTemp(dir, pattern string) (name string, err error)
}

// Remove is the interface implemented by a file system that supports removing files.
type Remove interface {
	FS

	// Remove removes the named file or (empty) directory.
	// If there is an error, it should be of type [*PathError].
	Remove(name string) error
}

// RemoveAll is the interface implemented by a file system that supports removing directories and their contents.
type RemoveAll interface {
	FS

	// RemoveAll removes path and any children it contains.
	// It removes everything it can but returns the first error
	// it encounters. If the path does not exist, RemoveAll
	// should return nil (no error).
	// If there is an error, it should be of type [*PathError].
	RemoveAll(name string) error
}

// WriteFile is the interface implemented by a file system that supports writing files.
type WriteFile interface {
	FS

	// WriteFile writes data to the named file.
	WriteFile(name string, data []byte, perm FileMode) error
}

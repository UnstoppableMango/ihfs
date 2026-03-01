// Package ihfs provides composable filesystem interfaces aligned with Go's io/fs package philosophy.
package ihfs

import (
	"io"
	"io/fs"
)

type (
	// DirEntry is an alias for [fs.DirEntry].
	DirEntry = fs.DirEntry
	// File is an alias for [fs.File].
	File = fs.File
	// FileInfo is an alias for [fs.FileInfo].
	FileInfo = fs.FileInfo
	// FileMode is an alias for [fs.FileMode].
	FileMode = fs.FileMode
	// PathError is an alias for [fs.PathError].
	PathError = fs.PathError
	// ReadDirFile is an alias for [fs.ReadDirFile].
	ReadDirFile = fs.ReadDirFile
)

var (
	// ErrClosed is an alias for [fs.ErrClosed].
	ErrClosed = fs.ErrClosed
	// ErrExist is an alias for [fs.ErrExist].
	ErrExist      = fs.ErrExist
	// ErrInvalid is an alias for [fs.ErrInvalid].
	ErrInvalid = fs.ErrInvalid
	// ErrNotExist is an alias for [fs.ErrNotExist].
	ErrNotExist = fs.ErrNotExist
	// ErrPermission is an alias for [fs.ErrPermission].
	ErrPermission = fs.ErrPermission
)

// DirReader is the interface implemented by a file that supports reading directory entries.
type DirReader interface {
	File

	// ReadDir reads the contents of the directory and returns
	// a slice of up to n DirEntry values in directory order.
	//
	// If n > 0, ReadDir returns at most n DirEntry values.
	// In this case, if ReadDir returns an empty slice, it will
	// return a non-nil error explaining why.
	//
	// If n <= 0, ReadDir returns all the DirEntry values from
	// the directory in a single slice. In this case, if
	// ReadDir succeeds (reads all entries), it returns a nil error.
	ReadDir(n int) ([]DirEntry, error)
}

// DirNameReader is the interface implemented by a file that supports an optimized version of reading directory entry names.
type DirNameReader interface {
	File

	// ReadDirNames reads the contents of the directory
	// and returns a slice of names of up to n entries
	// in directory order.
	//
	// If n > 0, ReadDirNames returns at most n names.
	// In this case, if ReadDirNames returns an empty slice, it will
	// return a non-nil error explaining why.
	//
	// If n <= 0, ReadDirNames returns all the names from
	// the directory in a single slice. In this case, if
	// ReadDirNames succeeds (reads all entries), it returns a nil error.
	ReadDirNames(n int) ([]string, error)
}

// Operation represents a file system operation.
type Operation interface {
	// Subject returns the subject of the operation, typically a file or directory path.
	Subject() string
}

// ReaderAt is the interface implemented by a file that supports reading at a specific offset.
type ReaderAt interface {
	File
	io.ReaderAt
}

// Seeker is the interface implemented by a file that supports seeking.
type Seeker interface {
	File
	io.Seeker
}

// StringWriter is the interface implemented by a file that supports writing strings.
type StringWriter interface {
	File
	io.StringWriter
}

// Syncer is the interface implemented by a file that supports syncing its contents to stable storage.
type Syncer interface {
	File
	Sync() error
}

// Truncater is the interface implemented by a file that supports truncating its size.
type Truncater interface {
	File
	Truncate(size int64) error
}

// Writer is the interface implemented by a file that supports writing.
type Writer interface {
	File
	io.Writer
}

// WriterAt is the interface implemented by a file that supports writing at a specific offset.
type WriterAt interface {
	File
	io.WriterAt
}

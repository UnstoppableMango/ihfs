package ihfs

import (
	"io"
	"io/fs"
)

type (
	DirEntry  = fs.DirEntry
	File      = fs.File
	FileInfo  = fs.FileInfo
	FileMode  = fs.FileMode
	PathError = fs.PathError
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

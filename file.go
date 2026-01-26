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

// Seeker is the interface implemented by a file that supports seeking.
type Seeker interface {
	File
	io.Seeker
}

// Writer is the interface implemented by a file that supports writing.
type Writer interface {
	File
	io.Writer
}

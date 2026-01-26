package ihfs

import (
	"io"
	"io/fs"
)

type (
	Seeker = io.Seeker

	DirEntry  = fs.DirEntry
	File      = fs.File
	FileInfo  = fs.FileInfo
	FileMode  = fs.FileMode
	PathError = fs.PathError
)

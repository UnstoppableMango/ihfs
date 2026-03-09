package ghfs

import (
	"io"
	"io/fs"
	"time"
)

type FileInfo struct {
	name  string
	rc    io.ReadCloser
	isDir bool
}

func (fi *FileInfo) Name() string       { return fi.name }
func (fi *FileInfo) IsDir() bool        { return fi.isDir }
func (fi *FileInfo) ModTime() time.Time { return time.Time{} }
func (fi *FileInfo) Sys() any           { return fi.rc }

func (fi *FileInfo) Mode() fs.FileMode {
	if fi.isDir {
		return fs.ModeDir | 0555
	}
	return 0444
}

func (fi *FileInfo) Size() int64 {
	if fi.isDir {
		return 0
	}
	return -1
}

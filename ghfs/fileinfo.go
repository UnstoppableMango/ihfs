package ghfs

import (
	"io"
	"io/fs"
	"time"
)

type FileInfo struct {
	name string
	rc   io.ReadCloser
}

func (fi *FileInfo) Name() string       { return fi.name }
func (fi *FileInfo) IsDir() bool        { return false }
func (fi *FileInfo) ModTime() time.Time { return time.Time{} }
func (fi *FileInfo) Mode() fs.FileMode  { return 0444 }
func (fi *FileInfo) Size() int64        { return -1 }
func (fi *FileInfo) Sys() any           { return fi.rc }

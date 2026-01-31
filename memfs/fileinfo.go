package memfs

import (
	"os"
	"time"

	"github.com/unstoppablemango/ihfs"
)

// FileInfo implements ihfs.FileInfo for in-memory files.
type FileInfo struct {
	data *FileData
}

// Name implements ihfs.FileInfo.
func (fi *FileInfo) Name() string {
	fi.data.Lock()
	defer fi.data.Unlock()
	return fi.data.name
}

// Size implements ihfs.FileInfo.
func (fi *FileInfo) Size() int64 {
	fi.data.Lock()
	defer fi.data.Unlock()

	if fi.data.isDir {
		return 0
	}
	return int64(len(fi.data.content))
}

// Mode implements ihfs.FileInfo.
func (fi *FileInfo) Mode() os.FileMode {
	fi.data.Lock()
	defer fi.data.Unlock()
	return fi.data.mode
}

// ModTime implements ihfs.FileInfo.
func (fi *FileInfo) ModTime() time.Time {
	fi.data.Lock()
	defer fi.data.Unlock()
	return fi.data.modTime
}

// IsDir implements ihfs.FileInfo.
func (fi *FileInfo) IsDir() bool {
	fi.data.Lock()
	defer fi.data.Unlock()
	return fi.data.isDir
}

// Sys implements ihfs.FileInfo.
func (fi *FileInfo) Sys() any {
	return fi.data
}

// Type implements ihfs.DirEntry.
func (fi *FileInfo) Type() os.FileMode {
	return fi.Mode().Type()
}

// Info implements ihfs.DirEntry.
func (fi *FileInfo) Info() (ihfs.FileInfo, error) {
	return fi, nil
}

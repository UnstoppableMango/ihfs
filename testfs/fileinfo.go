package testfs

import (
	"io/fs"
	"time"
)

// TODO: make this API less awkward to use in tests

// FileInfo is a configurable test [fs.FileInfo] implementation.
type FileInfo struct {
	NameV    string
	IsDirV   bool
	SizeV    int64
	ModeV    fs.FileMode
	ModTimeV time.Time
	SysV     any
}

// NewFileInfo creates a new [FileInfo] with the given name and default zero values.
func NewFileInfo(name string) *FileInfo {
	return &FileInfo{
		name:        name,
		IsDirFunc:   func() bool { return false },
		SizeFunc:    func() int64 { return 0 },
		ModeFunc:    func() fs.FileMode { return 0 },
		ModTimeFunc: func() time.Time { return time.Time{} },
		SysFunc:     func() any { return nil },
	}
}

// Name implements [fs.FileInfo].
func (fi *FileInfo) Name() string {
	return fi.name
}

// IsDir implements [fs.FileInfo].
func (fi *FileInfo) IsDir() bool {
	return fi.IsDirFunc()
}

// Size implements [fs.FileInfo].
func (fi *FileInfo) Size() int64 {
	return fi.SizeFunc()
}

// Mode implements [fs.FileInfo].
func (fi *FileInfo) Mode() fs.FileMode {
	return fi.ModeFunc()
}

// ModTime implements [fs.FileInfo].
func (fi *FileInfo) ModTime() time.Time {
	return fi.ModTimeFunc()
}

// Sys implements [fs.FileInfo].
func (fi *FileInfo) Sys() any {
	return fi.SysFunc()
}

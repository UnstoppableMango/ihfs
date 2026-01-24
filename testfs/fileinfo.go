package testfs

import (
	"io/fs"
	"time"
)

type FileInfo struct {
	NameFunc    func() string
	IsDirFunc   func() bool
	SizeFunc    func() int64
	ModeFunc    func() fs.FileMode
	ModTimeFunc func() time.Time
	SysFunc     func() any
}

func NewFileInfo() *FileInfo {
	return &FileInfo{
		NameFunc:    func() string { return "" },
		IsDirFunc:   func() bool { return false },
		SizeFunc:    func() int64 { return 0 },
		ModeFunc:    func() fs.FileMode { return 0 },
		ModTimeFunc: func() time.Time { return time.Time{} },
		SysFunc:     func() any { return nil },
	}
}

func (fi *FileInfo) Name() string {
	return fi.NameFunc()
}

func (fi *FileInfo) IsDir() bool {
	return fi.IsDirFunc()
}

func (fi *FileInfo) Size() int64 {
	return fi.SizeFunc()
}

func (fi *FileInfo) Mode() fs.FileMode {
	return fi.ModeFunc()
}

func (fi *FileInfo) ModTime() time.Time {
	return fi.ModTimeFunc()
}

func (fi *FileInfo) Sys() any {
	return fi.SysFunc()
}

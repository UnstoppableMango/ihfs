package ghfs

import (
	"encoding/json"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/unstoppablemango/ihfs"
)

type File struct {
	name  string
	rc    io.ReadCloser
	isDir bool
}

func (f *File) Close() error {
	if f.rc == nil {
		return nil
	}
	return f.rc.Close()
}

func (f *File) Read(p []byte) (n int, err error) {
	if f.isDir {
		return 0, f.error("read", fs.ErrInvalid)
	}
	if f.rc == nil {
		return 0, nil
	}
	return f.rc.Read(p)
}

func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {
	if !f.isDir {
		return nil, f.error("readdir", fs.ErrInvalid)
	}

	// TODO: implement ReadDir for directories
	if n > 0 {
		return nil, io.EOF
	}
	return nil, nil
}

func (f *File) Stat() (ihfs.FileInfo, error) {
	base, _, _ := strings.Cut(filepath.Base(f.name), "?")

	return &FileInfo{
		name:  base,
		rc:    f.rc,
		isDir: f.isDir,
	}, nil
}

func (f *File) Decode(v any) error {
	return json.NewDecoder(f).Decode(v)
}

func (f *File) error(op string, err error) error {
	return &fs.PathError{
		Op:   op,
		Path: f.name,
		Err:  err,
	}
}

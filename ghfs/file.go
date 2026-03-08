package ghfs

import (
	"encoding/json"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/unstoppablemango/ihfs"
)

type File struct {
	io.ReadCloser
	name string
}

func (f *File) IsDir() bool        { return false }
func (f *File) ModTime() time.Time { return time.Time{} }
func (f *File) Mode() fs.FileMode  { return 0444 }
func (f *File) Size() int64        { return -1 }
func (f *File) Sys() any           { return f.ReadCloser }

func (f *File) Close() error {
	if f.ReadCloser != nil {
		return f.ReadCloser.Close()
	}
	return nil
}

func (f *File) Name() string {
	base := filepath.Base(f.name)
	if name, _, ok := strings.Cut(base, "?"); ok {
		return name
	}
	return base
}

func (f *File) Stat() (ihfs.FileInfo, error) {
	return f, nil
}

func (f *File) Decode(v any) error {
	return json.NewDecoder(f).Decode(v)
}

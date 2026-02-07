package ghfs

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/unstoppablemango/ihfs"
)

type File struct {
	*bytes.Reader
	name string
}

func (f *File) IsDir() bool        { return false }
func (f *File) ModTime() time.Time { return time.Time{} }
func (f *File) Mode() fs.FileMode  { return 0444 }
func (f *File) Size() int64        { return int64(f.Len()) }
func (f *File) Sys() any           { return f.Reader }
func (f *File) Close() error       { return nil }

func (f *File) Name() string {
	return filepath.Base(f.name)
}

func (f *File) Stat() (ihfs.FileInfo, error) {
	return f, nil
}

func (f *File) Decode(v any) error {
	return json.NewDecoder(f).Decode(v)
}

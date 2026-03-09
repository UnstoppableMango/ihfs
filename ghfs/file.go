package ghfs

import (
	"encoding/json"
	"io"
	"path/filepath"
	"strings"

	"github.com/unstoppablemango/ihfs"
)

type File struct {
	name string
	rc   io.ReadCloser
}

func (f *File) Close() error {
	if f.rc == nil {
		return nil
	}
	return f.rc.Close()
}

func (f *File) Read(p []byte) (n int, err error) {
	if f.rc == nil {
		return 0, nil
	}
	return f.rc.Read(p)
}

func (f *File) Stat() (ihfs.FileInfo, error) {
	base, _, _ := strings.Cut(filepath.Base(f.name), "?")

	return &FileInfo{
		name: base,
		rc:   f.rc,
	}, nil
}

func (f *File) Decode(v any) error {
	return json.NewDecoder(f).Decode(v)
}

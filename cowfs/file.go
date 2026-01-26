package cowfs

import (
	"io"
	"io/fs"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/fsutil/try"
)

// File represents a file in the copy-on-write filesystem. It wraps a file from
// the base layer and a file from the overlay layer. Reads are served from the
// overlay if present, otherwise from the base. Writes are directed to the
// overlay.
type File struct {
	name  string
	base  ihfs.File
	layer ihfs.File
}

// Close implements [fs.File].
func (f *File) Close() error {
	// Base should be closed first so that the overlay has a newer
	// timestamp, otherwise the cache will never get hit.
	if f.base != nil {
		f.base.Close()
	}
	if f.layer != nil {
		return f.layer.Close()
	}
	return BADFD
}

// Read implements [fs.File].
func (f *File) Read(b []byte) (int, error) {
	if f.layer != nil {
		n, err := f.layer.Read(b)
		if (err == nil || err == io.EOF) && f.base != nil {
			o, w := int64(n), io.SeekCurrent
			if _, seekErr := try.Seek(f.base, o, w); seekErr != nil {
				err = seekErr
			}
		}
		return n, err
	}
	if f.base != nil {
		return f.base.Read(b)
	}

	return 0, BADFD
}

// Stat implements [fs.File].
func (f *File) Stat() (fs.FileInfo, error) {
	if f.layer != nil {
		return f.layer.Stat()
	}
	if f.base != nil {
		return f.base.Stat()
	}
	return nil, BADFD
}

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
	Base  ihfs.File
	Layer ihfs.File
}

func NewFile(base, layer ihfs.File) *File {
	return &File{Base: base, Layer: layer}
}

// Close implements [fs.File].
func (f *File) Close() error {
	// Base should be closed first so that the overlay has a newer
	// timestamp, otherwise the cache will never get hit.
	if f.Base != nil {
		f.Base.Close()
	}
	if f.Layer != nil {
		return f.Layer.Close()
	}
	return BADFD
}

// Read implements [fs.File].
func (f *File) Read(b []byte) (int, error) {
	if f.Layer != nil {
		n, err := f.Layer.Read(b)
		if (err == nil || err == io.EOF) && f.Base != nil {
			o, w := int64(n), io.SeekCurrent
			if _, seekErr := try.Seek(f.Base, o, w); seekErr != nil {
				err = seekErr
			}
		}
		return n, err
	}
	if f.Base != nil {
		return f.Base.Read(b)
	}

	return 0, BADFD
}

// Stat implements [fs.File].
func (f *File) Stat() (fs.FileInfo, error) {
	if f.Layer != nil {
		return f.Layer.Stat()
	}
	if f.Base != nil {
		return f.Base.Stat()
	}
	return nil, BADFD
}

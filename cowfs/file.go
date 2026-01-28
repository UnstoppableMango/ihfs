package cowfs

import (
	"errors"
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
	base  ihfs.File
	layer ihfs.File
}

// NewFile creates a new copy-on-write file with the given base and layer files.
func NewFile(base, layer ihfs.File) *File {
	return &File{base, layer}
}

// Close implements [fs.File].
func (f *File) Close() error {
	if f.base == nil && f.layer == nil {
		return BADFD
	}

	// Base should be closed first so that the overlay has a newer
	// timestamp, otherwise the cache would never get hit.
	var baseErr, layerErr error
	if f.base != nil {
		baseErr = f.base.Close()
	}
	if f.layer != nil {
		layerErr = f.layer.Close()
	}
	if baseErr != nil || layerErr != nil {
		return errors.Join(baseErr, layerErr)
	}
	return nil
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

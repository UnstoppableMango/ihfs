package v1

import (
	"io"

	"github.com/unstoppablemango/ihfs/tarfs"
)

// FS wraps a tar stream as a read-only file system.
//
// Call [FS.Close] when the FS is no longer needed to release the underlying stream.
type FS struct {
	*tarfs.Reader
	closer io.Closer
}

// Close releases the underlying stream.
func (f *FS) Close() error {
	return f.closer.Close()
}

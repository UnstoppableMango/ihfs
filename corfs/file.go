package corfs

import (
	"errors"
	"io"

	"github.com/unstoppablemango/ihfs"
)

// File represents a file in the cache-on-read filesystem.
// For directories, it wraps files from both base and layer.
// For regular files, it only wraps the cached layer file.
type File struct {
	base  ihfs.File
	layer ihfs.File
}

// NewFile creates a new cache-on-read file with the given base and layer files.
func NewFile(base, layer ihfs.File) *File {
	return newFile(base, layer)
}

func newFile(base, layer ihfs.File) *File {
	return &File{
		base:  base,
		layer: layer,
	}
}

// Close implements [fs.File].
func (f *File) Close() error {
	if f.base == nil && f.layer == nil {
		return BADFD
	}

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
		return f.layer.Read(b)
	}
	if f.base != nil {
		return f.base.Read(b)
	}
	return 0, BADFD
}

// Stat implements [fs.File].
func (f *File) Stat() (ihfs.FileInfo, error) {
	if f.layer != nil {
		return f.layer.Stat()
	}
	if f.base != nil {
		return f.base.Stat()
	}
	return nil, BADFD
}

// ReadDir reads the contents of the directory. For cache-on-read,
// directories are merged from both base and layer, similar to cowfs.
func (f *File) ReadDir(n int) ([]ihfs.DirEntry, error) {
	var entries []ihfs.DirEntry

	// Get entries from layer if available
	if f.layer != nil {
		if dir, ok := f.layer.(ihfs.ReadDirFile); ok {
			layerEntries, err := dir.ReadDir(-1)
			if err != nil {
				return nil, err
			}
			entries = append(entries, layerEntries...)
		}
	}

	// Get entries from base if available
	if f.base != nil {
		if dir, ok := f.base.(ihfs.ReadDirFile); ok {
			baseEntries, err := dir.ReadDir(-1)
			if err != nil {
				return nil, err
			}
			// Merge entries, avoiding duplicates
			seen := make(map[string]bool)
			for _, e := range entries {
				seen[e.Name()] = true
			}
			for _, e := range baseEntries {
				if !seen[e.Name()] {
					entries = append(entries, e)
				}
			}
		}
	}

	// Handle pagination
	if n <= 0 {
		return entries, nil
	}
	if len(entries) == 0 {
		return nil, io.EOF
	}
	if n > len(entries) {
		n = len(entries)
	}

	return entries[:n], nil
}

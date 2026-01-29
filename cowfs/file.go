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
	base    ihfs.File
	layer   ihfs.File
	off     int
	entries []ihfs.DirEntry
}

// NewFile creates a new copy-on-write file with the given base and layer files.
func NewFile(base, layer ihfs.File) *File {
	return &File{base: base, layer: layer}
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

// ReadDir reads the contents of the directory and returns a slice of
// DirEntry values. It merges entries from both the base and layer,
// with layer entries taking precedence over base entries with the same name.
//
// If n > 0, ReadDir returns at most n DirEntry structures.
// In this case, if ReadDir returns an empty slice, it will return
// a non-nil error explaining why.
// At the end of a directory, the error is io.EOF.
//
// If n <= 0, ReadDir returns all the DirEntry values from the directory
// in a single slice. In this case, if ReadDir succeeds (reads all the way
// to the end of the directory), it returns the slice and a nil error.
// If it encounters an error before the end of the directory,
// ReadDir returns the DirEntry list read until that point and a non-nil error.
func (f *File) ReadDir(n int) ([]ihfs.DirEntry, error) {
	if f.off == 0 {
		var layerEntries []ihfs.DirEntry
		if f.layer != nil {
			if dir, ok := f.layer.(fs.ReadDirFile); ok {
				entries, err := dir.ReadDir(-1)
				if err != nil {
					return nil, err
				}
				layerEntries = entries
			}
		}

		var baseEntries []ihfs.DirEntry
		if f.base != nil {
			if dir, ok := f.base.(fs.ReadDirFile); ok {
				entries, err := dir.ReadDir(-1)
				if err != nil {
					return nil, err
				}
				baseEntries = entries
			}
		}

		merged := mergeDirEntries(layerEntries, baseEntries)
		f.entries = append(f.entries, merged...)
	}

	entries := f.entries[f.off:]

	if n <= 0 {
		f.off += len(entries)
		return entries, nil
	}

	if len(entries) == 0 {
		return nil, io.EOF
	}

	if n > len(entries) {
		n = len(entries)
	}

	f.off += n
	return entries[:n], nil
}

// mergeDirEntries merges directory entries from the layer and base,
// with layer entries taking precedence over base entries with the same name.
func mergeDirEntries(layer, base []ihfs.DirEntry) []ihfs.DirEntry {
	entries := make(map[string]ihfs.DirEntry)

	for _, entry := range layer {
		entries[entry.Name()] = entry
	}

	for _, entry := range base {
		if _, exists := entries[entry.Name()]; !exists {
			entries[entry.Name()] = entry
		}
	}

	result := make([]ihfs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry)
	}

	return result
}

// Write implements [ihfs.Writer].
func (f *File) Write(b []byte) (int, error) {
	if f.layer != nil {
		n, err := try.Write(f.layer, b)
		if err == nil || f.base != nil {
			_, err = try.Write(f.base, b)
		}
		return n, err
	}
	if f.base != nil {
		return try.Write(f.base, b)
	}

	return 0, BADFD
}

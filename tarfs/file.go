package tarfs

import (
	"archive/tar"
	"bytes"
	"cmp"
	"io"
	"io/fs"
	"slices"
	"strings"
	"time"

	"github.com/unstoppablemango/ihfs"
)

// File represents a file in a tar archive.
type File struct {
	io.Reader
	hdr *tar.Header
}

// Close implements [fs.File].
func (f *File) Close() error {
	return nil
}

// Stat implements [fs.File].
func (f *File) Stat() (fs.FileInfo, error) {
	return f.hdr.FileInfo(), nil
}

// Name returns the name of the tar entry.
func (f *File) Name() string {
	return f.hdr.Name
}

// dirFile represents a directory in a tar archive (both root and subdirectories).
type dirFile struct {
	name         string
	cache        *cache
	readDirCount int
}

func newDirFile(name string, cache *cache) *dirFile {
	return &dirFile{name: name, cache: cache}
}

func (f *dirFile) Close() error {
	return nil
}

func (f *dirFile) Read(p []byte) (int, error) {
	return 0, &ihfs.PathError{Op: "read", Path: f.name, Err: fs.ErrInvalid}
}

func (f *dirFile) Stat() (fs.FileInfo, error) {
	return dirFileInfo{name: f.name}, nil
}

func (f *dirFile) ReadDir(n int) ([]ihfs.DirEntry, error) {
	// Determine prefix: empty for root ("."), otherwise name + "/"
	var prefix string
	if f.name != "." {
		prefix = f.name + "/"
	}

	// Collect entries under this directory
	var entries []ihfs.DirEntry
	seen := make(map[string]bool)

	for _, fd := range f.cache.all() {
		// Skip entries that don't match our prefix
		if prefix != "" && !strings.HasPrefix(fd.hdr.Name, prefix) {
			continue
		}

		// Get the relative path after the prefix
		rel := strings.TrimPrefix(fd.hdr.Name, prefix)
		parts := strings.SplitN(rel, "/", 2)
		baseName := parts[0]

		if baseName == "" || seen[baseName] {
			continue
		}
		seen[baseName] = true

		// If this is a subdirectory (has more parts), create a synthetic entry
		if len(parts) > 1 {
			entries = append(entries, dirEntry{name: baseName})
		} else {
			// It's a file directly under this directory
			entries = append(entries, fs.FileInfoToDirEntry(fd.hdr.FileInfo()))
		}
	}

	// Sort entries by name
	slices.SortFunc(entries, func(a, b ihfs.DirEntry) int {
		return cmp.Compare(a.Name(), b.Name())
	})

	if n <= 0 {
		result := entries[f.readDirCount:]
		f.readDirCount = len(entries)
		return result, nil
	}

	// Return n entries
	start := f.readDirCount
	if start >= len(entries) {
		return nil, io.EOF
	}

	end := start + n
	if end > len(entries) {
		end = len(entries)
	}

	result := entries[start:end]
	f.readDirCount = end

	return result, nil
}

// dirEntry is a synthetic directory entry for intermediate directories
type dirEntry struct {
	name string
}

func (d dirEntry) Name() string               { return d.name }
func (d dirEntry) IsDir() bool                { return true }
func (d dirEntry) Type() fs.FileMode          { return fs.ModeDir }
func (d dirEntry) Info() (fs.FileInfo, error) { return dirFileInfo{d.name}, nil }

type dirFileInfo struct {
	name string
}

func (dfi dirFileInfo) Name() string       { return dfi.name }
func (dfi dirFileInfo) Size() int64        { return 0 }
func (dfi dirFileInfo) Mode() fs.FileMode  { return fs.ModeDir | 0755 }
func (dfi dirFileInfo) ModTime() time.Time { return time.Time{} }
func (dfi dirFileInfo) IsDir() bool        { return true }
func (dfi dirFileInfo) Sys() interface{}   { return nil }

type fileData struct {
	hdr  *tar.Header
	data []byte
}

func (fd fileData) file() *File {
	return &File{
		hdr:    fd.hdr,
		Reader: bytes.NewReader(fd.data),
	}
}

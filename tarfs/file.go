package tarfs

import (
	"archive/tar"
	"bytes"
	"cmp"
	"io"
	"io/fs"
	"path"
	"slices"
	"strings"
	"time"
)

// File represents a file in a tar archive.
type File struct {
	io.Reader
	hdr          *tar.Header
	cache        *cache
	name         string
	readDirCount int
}

// Close implements [fs.File].
func (f *File) Close() error {
	return nil
}

// Read implements [io.Reader]. For directories, returns an error.
func (f *File) Read(p []byte) (int, error) {
	// For directories, return error (cannot read directory content)
	if f.hdr.FileInfo().IsDir() {
		return 0, &fs.PathError{Op: "read", Path: f.name, Err: fs.ErrInvalid}
	}
	return f.Reader.Read(p)
}

// Stat implements [fs.File].
func (f *File) Stat() (fs.FileInfo, error) {
	// For synthetic directories (created by us, not from tar), return FileInfo with nil Sys()
	if f.hdr.Typeflag == tar.TypeDir && f.hdr.Size == 0 && f.cache != nil {
		// Check if this is a synthetic directory (not actually in the tar)
		if f.cache.get(f.name) == nil {
			return fileInfo{hdr: f.hdr, nilSys: true}, nil
		}
	}
	return f.hdr.FileInfo(), nil
}

// Name returns the name of the tar entry.
func (f *File) Name() string {
	return f.hdr.Name
}

// ReadDir implements [fs.ReadDirFile] for directories.
func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {
	if !f.hdr.FileInfo().IsDir() {
		return nil, &fs.PathError{Op: "readdir", Path: f.name, Err: fs.ErrInvalid}
	}

	// Determine prefix: empty for root ("."), otherwise name + "/"
	var prefix string
	if f.name != "." {
		prefix = f.name + "/"
	}

	// Collect entries under this directory
	var entries []fs.DirEntry
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
			entries = append(entries, fileInfo{name: baseName})
		} else {
			// It's a file directly under this directory
			entries = append(entries, fs.FileInfoToDirEntry(fd.hdr.FileInfo()))
		}
	}

	// Sort entries by name
	slices.SortFunc(entries, func(a, b fs.DirEntry) int {
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

// fileInfo wraps tar.Header as fs.FileInfo and fs.DirEntry, or represents a synthetic directory by name
type fileInfo struct {
	hdr    *tar.Header
	name   string // used when hdr is nil (for synthetic subdirectories)
	nilSys bool   // when true, Sys() returns nil
}

func (fi fileInfo) Name() string {
	if fi.hdr != nil {
		return path.Base(fi.hdr.Name)
	}
	return path.Base(fi.name)
}

func (fi fileInfo) Size() int64 { return 0 }

func (fi fileInfo) Mode() fs.FileMode {
	if fi.hdr != nil {
		return fs.ModeDir | fs.FileMode(fi.hdr.Mode)
	}
	return fs.ModeDir | 0755
}

func (fi fileInfo) ModTime() time.Time {
	if fi.hdr != nil {
		return fi.hdr.ModTime
	}
	return time.Time{}
}

func (fi fileInfo) IsDir() bool { return true }

func (fi fileInfo) Sys() interface{} {
	if fi.nilSys || fi.hdr == nil {
		return nil
	}
	return fi.hdr
}

// Type implements [fs.DirEntry].
func (fi fileInfo) Type() fs.FileMode {
	return fs.ModeDir
}

// Info implements [fs.DirEntry].
func (fi fileInfo) Info() (fs.FileInfo, error) {
	return fi, nil
}

type fileData struct {
	hdr  *tar.Header
	data []byte
}

func (fd fileData) file(cache *cache) *File {
	return &File{
		hdr:    fd.hdr,
		name:   fd.hdr.Name,
		cache:  cache,
		Reader: bytes.NewReader(fd.data),
	}
}

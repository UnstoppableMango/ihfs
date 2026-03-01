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
)

// File represents a file in a tar archive.
type File struct {
	r               io.Reader
	hdr             *tar.Header
	cache           *cache
	name            string
	readDirCount    int
	readDirSnapshot []fs.DirEntry // stable snapshot built on first ReadDir call
}

// Close implements [fs.File].
func (f *File) Close() error {
	return nil
}

// FileInfo returns the [fs.FileInfo] for the tar entry.
func (f *File) FileInfo() fs.FileInfo {
	return f.hdr.FileInfo()
}

// IsDir reports whether the tar entry is a directory.
func (f *File) IsDir() bool {
	return f.FileInfo().IsDir()
}

// Read implements [io.Reader]. For directories, returns an error.
func (f *File) Read(p []byte) (int, error) {
	if f.IsDir() {
		return 0, f.perror("read", fs.ErrInvalid)
	}
	return f.r.Read(p)
}

// Stat implements [fs.File].
func (f *File) Stat() (fs.FileInfo, error) {
	return f.FileInfo(), nil
}

// Name returns the name of the tar entry.
func (f *File) Name() string {
	return f.hdr.Name
}

// ReadDir implements [fs.ReadDirFile] for directories.
func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {
	if !f.IsDir() {
		return nil, f.perror("readdir", fs.ErrInvalid)
	}

	// Snapshot on first call so paginated reads are stable.
	if f.readDirSnapshot == nil {
		f.readDirSnapshot = f.buildDirSnapshot()
	}
	entries := f.readDirSnapshot

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

	end := min(start+n, len(entries))
	result := entries[start:end]
	f.readDirCount = end

	return result, nil
}

// buildDirSnapshot collects and sorts the directory entries visible in the cache
// at the time of the call. The result is stored as an immutable snapshot so that
// concurrent or sequential ReadDir calls paginate over a consistent view.
func (f *File) buildDirSnapshot() []fs.DirEntry {
	// Determine prefix: empty for root ("."), otherwise name + "/"
	var prefix string
	if f.name != "." {
		prefix = f.name + "/"
	}

	entries := make([]fs.DirEntry, 0)
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

		if len(parts) > 1 {
			// Subdirectory - prefer a real directory entry if cached, else synthetic
			dirKey := prefix + baseName
			if realEntry := f.cache.get(dirKey); realEntry != nil && realEntry.hdr.FileInfo().IsDir() {
				entries = append(entries, realEntry.dirEntry())
			} else {
				entries = append(entries, DirEntry{hdr: &tar.Header{
					Name:     baseName,
					Typeflag: tar.TypeDir,
					Mode:     0755,
				}, name: baseName})
			}
		} else {
			// It's a file directly under this directory
			entries = append(entries, fd.dirEntry())
		}
	}

	slices.SortFunc(entries, func(a, b fs.DirEntry) int {
		return cmp.Compare(a.Name(), b.Name())
	})

	return entries
}

func (f *File) perror(op string, err error) error {
	return &fs.PathError{Op: op, Path: f.name, Err: err}
}

// DirEntry represents a directory entry in a tar archive.
type DirEntry struct {
	hdr  *tar.Header
	name string
}

// Info implements [fs.DirEntry].
func (d DirEntry) Info() (fs.FileInfo, error) {
	return d.fileInfo(), nil
}

// IsDir implements [fs.DirEntry].
func (d DirEntry) IsDir() bool {
	return d.fileInfo().IsDir()
}

// Name implements [fs.DirEntry].
func (d DirEntry) Name() string {
	return path.Base(d.name)
}

// Type implements [fs.DirEntry].
func (d DirEntry) Type() fs.FileMode {
	return d.fileInfo().Mode().Type()
}

func (d DirEntry) fileInfo() fs.FileInfo {
	return d.hdr.FileInfo()
}

type fileData struct {
	hdr  *tar.Header
	data []byte
}

func (fd fileData) dirEntry() fs.DirEntry {
	return fs.FileInfoToDirEntry(fd.fileInfo())
}

func (fd fileData) fileInfo() fs.FileInfo {
	return fd.hdr.FileInfo()
}

func (fd fileData) file(cache *cache) *File {
	// Trim trailing slash from directory names so ReadDir computes the correct
	// prefix ("dir/") rather than a double-slash prefix ("dir//").
	name := fd.hdr.Name
	if fd.hdr.Typeflag == tar.TypeDir {
		name = strings.TrimSuffix(name, "/")
	}
	return &File{
		hdr:   fd.hdr,
		name:  name,
		cache: cache,
		r:     bytes.NewReader(fd.data),
	}
}

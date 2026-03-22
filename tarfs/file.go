package tarfs

import (
	"archive/tar"
	"bytes"
	"cmp"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"sync"
)

// File represents a file in a tar archive.
type File struct {
	name string
	hdr  *tar.Header
	r    io.Reader
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
	return false
}

// Read implements [io.Reader]. For directories, returns an error.
func (f *File) Read(p []byte) (int, error) {
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

func (f *File) store(m *sync.Map) {
	m.Store(f.name, f)
}

func (f *File) within(name string) bool {
	return strings.HasPrefix(name, f.name+"/")
}

func (f *File) dirEntry() fs.DirEntry {
	return &DirEntry{
		name: f.name,
	}
}

// TarError represents an error that occurred while accessing a file in a tar archive.
type TarError struct {
	Archive, Name string
	Err, Cause    error
}

func (e *TarError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf(
			"%s(%s): %v: %v",
			e.Archive, e.Name, e.Err, e.Cause,
		)
	}
	return fmt.Sprintf("%s(%s): %v", e.Archive, e.Name, e.Err)
}

func (e *TarError) Unwrap() []error {
	return []error{e.Err, e.Cause}
}

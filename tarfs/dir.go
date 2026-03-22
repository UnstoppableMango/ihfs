package tarfs

import (
	"archive/tar"
	"io/fs"
	"path"
	"strings"
	"sync"
	"time"
)

type Dir struct {
	name    string
	fsys    *Fs
	entries []fs.DirEntry
	index   int
}

func (d *Dir) Close() error { return nil }

// Read implements [fs.File].
func (d *Dir) Read([]byte) (int, error) {
	return 0, d.error("read", fs.ErrInvalid)
}

// Stat implements [fs.File].
func (d *Dir) Stat() (fs.FileInfo, error) {
	return &DirInfo{name: d.name}, nil
}

// ReadDir implements [fs.ReadDirFile].
func (d *Dir) ReadDir(n int) (entries []fs.DirEntry, err error) {
	if d.entries == nil {
		if d.entries, err = d.fsys.ReadDir(d.name); err != nil {
			return nil, err
		}
	}
	if d.index >= len(d.entries) {
		return nil, nil
	}
	if n <= 0 {
		return d.entries[d.index:], nil
	}

	entries = d.entries[d.index:]
	if n < len(entries) {
		entries = entries[:n]
	}

	d.index += len(entries)
	return entries, nil
}

func (d *Dir) store(m *sync.Map) {
	m.Store(d.name, d)
}

func (d *Dir) dirEntry() fs.DirEntry {
	return &DirEntry{
		name: d.name,
	}
}

func (d *Dir) within(name string) bool {
	return strings.HasPrefix(name, d.name+"/")
}

func (d *Dir) error(op string, err error) error {
	return &fs.PathError{Op: op, Path: d.name, Err: err}
}

type DirInfo struct {
	name string
	mod  time.Time
	size int64
}

func (d *DirInfo) IsDir() bool        { return true }
func (d *DirInfo) ModTime() time.Time { return d.mod }
func (d *DirInfo) Mode() fs.FileMode  { return fs.ModeDir }
func (d *DirInfo) Name() string       { return d.name }
func (d *DirInfo) Size() int64        { return d.size }
func (d *DirInfo) Sys() any           { return d }

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

func dirEntries(m *sync.Map, dir string) (entries []fs.DirEntry) {
	m.Range(func(key, value any) bool {
		if e, ok := value.(entry); ok && e.within(dir) {
			entries = append(entries, e.dirEntry())
		}
		return true
	})
	return
}

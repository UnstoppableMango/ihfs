package memfs

import (
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/unstoppablemango/ihfs"
)

// File represents a file in the memory filesystem.
type File struct {
	at           int64
	readDirCount int64
	closed       bool
	readOnly     bool
	data         *FileData
}

// FileData holds the actual file data and metadata.
type FileData struct {
	sync.Mutex
	name    string
	content []byte
	dir     *Dir
	isDir   bool
	mode    os.FileMode
	modTime time.Time
	uid     int
	gid     int
}

func (fd *FileData) error(op string, err error) error {
	return &ihfs.PathError{
		Op:   op,
		Path: fd.name,
		Err:  err,
	}
}

// Dir represents a directory with its children.
type Dir struct {
	sync.Mutex
	children map[string]*FileData
}

// NewFile creates a new file handle for the given file data.
func NewFile(data *FileData) *File {
	return &File{data: data}
}

// NewReadOnlyFile creates a new read-only file handle.
func NewReadOnlyFile(data *FileData) *File {
	return &File{data: data, readOnly: true}
}

// CreateFile creates new file data with the given name.
func CreateFile(name string) *FileData {
	return &FileData{
		name:    name,
		content: []byte{},
		mode:    os.ModeTemporary,
		modTime: time.Now(),
	}
}

// CreateDir creates new directory data with the given name.
func CreateDir(name string) *FileData {
	return &FileData{
		name:    name,
		dir:     &Dir{children: make(map[string]*FileData)},
		isDir:   true,
		mode:    os.ModeDir | 0755,
		modTime: time.Now(),
	}
}

// Close implements ihfs.File.
func (f *File) Close() error {
	f.data.Lock()
	defer f.data.Unlock()

	f.closed = true
	if !f.readOnly {
		f.data.modTime = time.Now()
	}
	return nil
}

// Read implements ihfs.File.
func (f *File) Read(p []byte) (int, error) {
	f.data.Lock()
	defer f.data.Unlock()

	if f.closed {
		return 0, ihfs.ErrClosed
	}

	if f.data.isDir {
		return 0, f.data.error("read", os.ErrInvalid)
	}

	at := atomic.LoadInt64(&f.at)
	if at >= int64(len(f.data.content)) {
		return 0, io.EOF
	}

	n := copy(p, f.data.content[at:])
	atomic.AddInt64(&f.at, int64(n))
	return n, nil
}

// Stat implements ihfs.File.
func (f *File) Stat() (ihfs.FileInfo, error) {
	return &FileInfo{data: f.data}, nil
}

// Write implements io.Writer.
func (f *File) Write(p []byte) (int, error) {
	if f.readOnly {
		return 0, f.data.error("write", os.ErrPermission)
	}

	f.data.Lock()
	defer f.data.Unlock()

	if f.closed {
		return 0, ihfs.ErrClosed
	}

	if f.data.isDir {
		return 0, f.data.error("write", os.ErrInvalid)
	}

	at := atomic.LoadInt64(&f.at)

	// Expand content if necessary
	if at > int64(len(f.data.content)) {
		f.data.content = append(f.data.content, make([]byte, at-int64(len(f.data.content)))...)
	}

	// Overwrite or append
	if at+int64(len(p)) > int64(len(f.data.content)) {
		f.data.content = append(f.data.content[:at], p...)
	} else {
		copy(f.data.content[at:], p)
	}

	atomic.AddInt64(&f.at, int64(len(p)))
	f.data.modTime = time.Now()

	return len(p), nil
}

// ReadDir implements fs.ReadDirFile.
func (f *File) ReadDir(n int) ([]ihfs.DirEntry, error) {
	f.data.Lock()
	defer f.data.Unlock()

	if f.closed {
		return nil, ihfs.ErrClosed
	}

	if !f.data.isDir {
		return nil, f.data.error("readdir", os.ErrInvalid)
	}

	f.data.dir.Lock()
	defer f.data.dir.Unlock()

	var entries []ihfs.DirEntry
	for _, child := range f.data.dir.children {
		entries = append(entries, &FileInfo{data: child})
	}

	sortDirEntries(entries)

	count := atomic.LoadInt64(&f.readDirCount)
	if n <= 0 {
		// Return all remaining entries
		if count >= int64(len(entries)) {
			return nil, io.EOF
		}
		result := entries[count:]
		atomic.StoreInt64(&f.readDirCount, int64(len(entries)))
		return result, nil
	}

	// Return n entries
	start := int(count)
	if start >= len(entries) {
		return nil, io.EOF
	}

	end := min(start+n, len(entries))
	atomic.StoreInt64(&f.readDirCount, int64(end))
	return entries[start:end], nil
}

// Seek implements io.Seeker.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	f.data.Lock()
	defer f.data.Unlock()

	if f.closed {
		return 0, ihfs.ErrClosed
	}

	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = atomic.LoadInt64(&f.at) + offset
	case io.SeekEnd:
		newPos = int64(len(f.data.content)) + offset
	default:
		return 0, f.error("seek", os.ErrInvalid)
	}

	if newPos < 0 {
		return 0, f.error("seek", os.ErrInvalid)
	}

	atomic.StoreInt64(&f.at, newPos)
	return newPos, nil
}

// Truncate implements ihfs truncation.
func (f *File) Truncate(size int64) error {
	if f.readOnly {
		return f.error("truncate", os.ErrPermission)
	}

	f.data.Lock()
	defer f.data.Unlock()

	if f.closed {
		return ihfs.ErrClosed
	}

	if size < 0 {
		return f.error("truncate", os.ErrInvalid)
	}

	if size > int64(len(f.data.content)) {
		// Extend with zeros
		f.data.content = append(f.data.content, make([]byte, size-int64(len(f.data.content)))...)
	} else {
		f.data.content = f.data.content[:size]
	}

	f.data.modTime = time.Now()
	return nil
}

// Sync implements file synchronization (no-op for in-memory).
func (f *File) Sync() error {
	return nil
}

func (f *File) error(op string, err error) error {
	return f.data.error(op, err)
}

func sortDirEntries(entries []ihfs.DirEntry) {
	// Simple bubble sort by name
	n := len(entries)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if entries[j].Name() > entries[j+1].Name() {
				entries[j], entries[j+1] = entries[j+1], entries[j]
			}
		}
	}
}

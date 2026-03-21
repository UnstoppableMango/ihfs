package tarfs

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"sync"
	"time"

	"github.com/unstoppablemango/ihfs"
)

// Writer provides a write-only tar-backed filesystem.
// Call [Writer.Close] to finalize the archive.
type Writer struct {
	name string
	tw   *tar.Writer
	mu   sync.Mutex
}

// NewWriter creates a Writer that writes a tar archive to w.
func NewWriter(name string, w io.Writer) *Writer {
	return &Writer{name: name, tw: tar.NewWriter(w)}
}

// Close finalizes the tar archive.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.tw.Close()
}

// Open implements [ihfs.FS]. Writer is write-only; Open always returns [ihfs.ErrPermission].
func (w *Writer) Open(name string) (ihfs.File, error) {
	return nil, &fs.PathError{Op: "open", Path: name, Err: ihfs.ErrPermission}
}

// Create implements [ihfs.CreateFS].
// The returned file buffers writes and flushes them as a tar entry on Close.
func (w *Writer) Create(name string) (ihfs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "create", Path: name, Err: ihfs.ErrInvalid}
	}
	return &writerFile{name: name, perm: 0666, w: w}, nil
}

// Mkdir implements [ihfs.MkdirFS].
func (w *Writer) Mkdir(name string, perm fs.FileMode) error {
	if !fs.ValidPath(name) {
		return &fs.PathError{Op: "mkdir", Path: name, Err: ihfs.ErrInvalid}
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.tw.WriteHeader(&tar.Header{
		Name:     name + "/",
		Typeflag: tar.TypeDir,
		Mode:     int64(perm),
		ModTime:  time.Now(),
	})
}

// WriteFile implements [ihfs.WriteFileFS].
func (w *Writer) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if !fs.ValidPath(name) {
		return &fs.PathError{Op: "writefile", Path: name, Err: ihfs.ErrInvalid}
	}
	return w.writeEntry(name, data, perm)
}

func (w *Writer) writeEntry(name string, data []byte, perm fs.FileMode) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.tw.WriteHeader(&tar.Header{
		Name:    name,
		Mode:    int64(perm),
		Size:    int64(len(data)),
		ModTime: time.Now(),
	})
	if err != nil {
		return err
	}

	_, err = w.tw.Write(data)
	return err
}

// writerFile is a buffered file handle that flushes content to the tar archive on Close.
type writerFile struct {
	name string
	perm fs.FileMode
	buf  bytes.Buffer
	w    *Writer
}

// Close flushes the buffered content as a tar entry.
func (f *writerFile) Close() error {
	return f.w.writeEntry(f.name, f.buf.Bytes(), f.perm)
}

// Read implements [fs.File]. writerFile is write-only; Read always returns [ihfs.ErrPermission].
func (f *writerFile) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: f.name, Err: ihfs.ErrPermission}
}

// Stat implements [fs.File]. writerFile is write-only; Stat always returns [ihfs.ErrPermission].
func (f *writerFile) Stat() (fs.FileInfo, error) {
	return nil, &fs.PathError{Op: "stat", Path: f.name, Err: ihfs.ErrPermission}
}

// Write implements [io.Writer].
func (f *writerFile) Write(p []byte) (int, error) {
	return f.buf.Write(p)
}

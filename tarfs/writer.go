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
	tw *tar.Writer
	mu sync.Mutex
}

// NewWriter creates a Writer that writes a tar archive to w.
// If w is already a [tar.Writer], it is used directly.
func NewWriter(w io.Writer) *Writer {
	tw, ok := w.(*tar.Writer)
	if !ok {
		tw = tar.NewWriter(w)
	}
	return &Writer{tw: tw}
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
	if err := w.tw.WriteHeader(&tar.Header{
		Name:     name + "/",
		Typeflag: tar.TypeDir,
		Mode:     int64(perm),
		ModTime:  time.Now(),
	}); err != nil {
		return &fs.PathError{Op: "mkdir", Path: name, Err: err}
	}
	return nil
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

	if err := w.tw.WriteHeader(&tar.Header{
		Name:    name,
		Mode:    int64(perm),
		Size:    int64(len(data)),
		ModTime: time.Now(),
	}); err != nil {
		return &fs.PathError{Op: "write", Path: name, Err: err}
	}

	if _, err := w.tw.Write(data); err != nil {
		return &fs.PathError{Op: "write", Path: name, Err: err}
	}
	return nil
}

// writerFile is a buffered file handle that flushes content to the tar archive on Close.
type writerFile struct {
	name   string
	perm   fs.FileMode
	buf    bytes.Buffer
	w      *Writer
	closed bool
}

func (f *writerFile) Name() string {
	return f.name
}

// Close flushes the buffered content as a tar entry.
// Subsequent calls are no-ops and return nil.
func (f *writerFile) Close() error {
	if f.closed {
		return nil
	}
	f.closed = true
	return f.w.writeEntry(f.name, f.buf.Bytes(), f.perm)
}

// Read implements [fs.File]. writerFile is write-only; Read always returns [ihfs.ErrPermission].
func (f *writerFile) Read([]byte) (int, error) {
	return 0, f.perror("read", ihfs.ErrPermission)
}

// Stat implements [fs.File]. writerFile is write-only; Stat always returns [ihfs.ErrPermission].
func (f *writerFile) Stat() (fs.FileInfo, error) {
	return nil, f.perror("stat", ihfs.ErrPermission)
}

// Write implements [io.Writer].
func (f *writerFile) Write(p []byte) (int, error) {
	return f.buf.Write(p)
}

func (f *writerFile) perror(op string, err error) error {
	return &fs.PathError{Op: op, Path: f.name, Err: err}
}

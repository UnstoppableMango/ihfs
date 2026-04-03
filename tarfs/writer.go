package tarfs

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/unstoppablemango/ihfs"
)

// Writer provides a write-only tar-backed filesystem.
// Call [Writer.Close] to finalize the archive.
type Writer struct {
	tw      *tar.Writer
	mu      sync.Mutex
	written map[string]struct{}
}

// NewWriter creates a Writer that writes a tar archive to w.
// If w is already a [tar.Writer], it is used directly.
func NewWriter(w io.Writer) *Writer {
	tw, ok := w.(*tar.Writer)
	if !ok {
		tw = tar.NewWriter(w)
	}
	return &Writer{
		tw:      tw,
		written: make(map[string]struct{}),
	}
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
	hdr := &tar.Header{
		Name:     name + "/",
		Typeflag: tar.TypeDir,
		Mode:     int64(perm.Perm()),
		ModTime:  time.Now(),
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.writeEntry(hdr, nil); err != nil {
		return &fs.PathError{Op: "mkdir", Path: name, Err: err}
	}
	return nil
}

// MkdirAll implements [ihfs.MkdirAllFS].
// It writes a directory tar entry for each path component not previously created.
// If name is already a directory, MkdirAll does nothing and returns nil.
func (w *Writer) MkdirAll(name string, perm fs.FileMode) error {
	if name == "." {
		return nil
	}
	if !fs.ValidPath(name) {
		return &fs.PathError{Op: "mkdirall", Path: name, Err: ihfs.ErrInvalid}
	}
	parts := strings.Split(name, "/")
	w.mu.Lock()
	defer w.mu.Unlock()
	for i := range parts {
		part := strings.Join(parts[:i+1], "/")
		hdr := &tar.Header{
			Name:     part + "/",
			Typeflag: tar.TypeDir,
			Mode:     int64(perm.Perm()),
			ModTime:  time.Now(),
		}
		if err := w.writeEntry(hdr, nil); err != nil {
			if errors.Is(err, fs.ErrExist) {
				continue
			}
			return &fs.PathError{Op: "mkdirall", Path: part, Err: err}
		}
	}
	return nil
}

// OpenFile implements [ihfs.OpenFileFS].
// Writer is write-only; OpenFile requires [os.O_WRONLY] or [os.O_RDWR] in flag.
func (w *Writer) OpenFile(name string, flag int, perm fs.FileMode) (ihfs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "openfile", Path: name, Err: ihfs.ErrInvalid}
	}
	if flag&(os.O_WRONLY|os.O_RDWR) == 0 {
		return nil, &fs.PathError{Op: "openfile", Path: name, Err: ihfs.ErrPermission}
	}
	return &writerFile{name: name, perm: perm, w: w}, nil
}

// Copy implements [ihfs.CopyFS].
// It walks fsys from the root and writes each entry into the archive under dir,
// preserving full metadata via [tar.FileInfoHeader].
func (w *Writer) Copy(dir string, fsys ihfs.FS) error {
	if dir != "" && !fs.ValidPath(dir) {
		return &fs.PathError{Op: "copy", Path: dir, Err: ihfs.ErrInvalid}
	}
	return fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := path.Join(dir, p)

		info, err := d.Info()
		if err != nil {
			return err
		}

		var link string
		if info.Mode()&fs.ModeSymlink != 0 {
			if link, err = fs.ReadLink(fsys, p); err != nil {
				return err
			}
		}

		hdr, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		hdr.Name = name
		if d.IsDir() && name != "." {
			hdr.Name += "/"
		}

		var r io.Reader
		if info.Mode().IsRegular() {
			f, err := fsys.Open(p)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			r = f
		}

		w.mu.Lock()
		defer w.mu.Unlock()
		if err := w.writeEntry(hdr, r); err != nil {
			if errors.Is(err, fs.ErrExist) {
				if d.IsDir() {
					return nil
				}
				return &fs.PathError{Op: "copy", Path: name, Err: fs.ErrExist}
			}
			return err
		}
		return nil
	})
}

// WriteEntry writes hdr and the optional content from r to the archive.
// r may be nil for entries with no body (directories, symlinks).
// Use [tar.FileInfoHeader] to build hdr from an [fs.FileInfo] to preserve metadata.
func (w *Writer) WriteEntry(hdr *tar.Header, r io.Reader) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writeEntry(hdr, r)
}

// writeEntry writes hdr and r to tw. Caller must hold w.mu.
// Returns [fs.ErrExist] if an entry with the same name was already written.
func (w *Writer) writeEntry(hdr *tar.Header, r io.Reader) error {
	key := strings.TrimSuffix(hdr.Name, "/")
	if _, seen := w.written[key]; seen {
		return fs.ErrExist
	}
	if err := w.tw.WriteHeader(hdr); err != nil {
		return err
	}
	if r != nil {
		if _, err := io.Copy(w.tw, r); err != nil {
			return err
		}
	}
	w.written[key] = struct{}{}
	return nil
}

// Symlink writes a symlink tar entry named newname pointing to oldname.
func (w *Writer) Symlink(oldname, newname string) error {
	if !fs.ValidPath(newname) {
		return &fs.PathError{Op: "symlink", Path: newname, Err: ihfs.ErrInvalid}
	}
	return w.WriteEntry(&tar.Header{
		Typeflag: tar.TypeSymlink,
		Name:     newname,
		Linkname: oldname,
		ModTime:  time.Now(),
	}, nil)
}

// WriteFile implements [ihfs.WriteFileFS].
func (w *Writer) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if !fs.ValidPath(name) {
		return &fs.PathError{Op: "writefile", Path: name, Err: ihfs.ErrInvalid}
	}
	err := w.WriteEntry(&tar.Header{
		Name:    name,
		Mode:    int64(perm.Perm()),
		Size:    int64(len(data)),
		ModTime: time.Now(),
	}, bytes.NewReader(data))
	if err != nil {
		return &fs.PathError{Op: "writefile", Path: name, Err: err}
	}
	return nil
}

// writerFile is a buffered file handle that flushes content to the tar archive on Close.
type writerFile struct {
	name     string
	perm     fs.FileMode
	buf      bytes.Buffer
	w        *Writer
	closed   bool
	closeErr error
}

func (f *writerFile) Name() string {
	return f.name
}

func (f *writerFile) Close() error {
	if f.closed {
		return f.closeErr
	}
	f.closed = true
	err := f.w.WriteEntry(&tar.Header{
		Name:    f.name,
		Mode:    int64(f.perm.Perm()),
		Size:    int64(f.buf.Len()),
		ModTime: time.Now(),
	}, &f.buf)
	if err != nil {
		f.closeErr = f.perror("close", err)
	}
	return f.closeErr
}

func (f *writerFile) Read([]byte) (int, error) {
	return 0, f.perror("read", ihfs.ErrPermission)
}

func (f *writerFile) Stat() (fs.FileInfo, error) {
	return nil, f.perror("stat", ihfs.ErrPermission)
}

// Write implements [io.Writer].
func (f *writerFile) Write(p []byte) (int, error) {
	if f.closed {
		return 0, f.perror("write", ihfs.ErrClosed)
	}
	return f.buf.Write(p)
}

func (f *writerFile) perror(op string, err error) error {
	return &fs.PathError{Op: op, Path: f.name, Err: err}
}

package try

import (
	"fmt"

	"github.com/unstoppablemango/ihfs"
)

// ReadAt attempts to call ReadAt on the given File.
// If the File does not implement [ihfs.ReaderAt], ReadAt returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func ReadAt(f ihfs.File, p []byte, off int64) (int, error) {
	if readerAt, ok := f.(ihfs.ReaderAt); ok {
		return readerAt.ReadAt(p, off)
	}
	return 0, fmt.Errorf("read at: %w", ErrNotImplemented)
}

// ReadDirFile attempts to call ReadDir on the given File.
// If the File does not implement [ihfs.DirReader], ReadDirFile returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func ReadDirFile(f ihfs.File, n int) ([]ihfs.DirEntry, error) {
	if dirReader, ok := f.(ihfs.DirReader); ok {
		return dirReader.ReadDir(n)
	}
	return nil, fmt.Errorf("read dir: %w", ErrNotImplemented)
}

// ReadDirNamesFile attempts to call ReadDirNames on the given File.
// If the File does not implement [ihfs.DirNameReader], ReadDirNamesFile returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func ReadDirNamesFile(f ihfs.File, n int) ([]string, error) {
	if dirNameReader, ok := f.(ihfs.DirNameReader); ok {
		return dirNameReader.ReadDirNames(n)
	}
	return nil, fmt.Errorf("read dir names: %w", ErrNotImplemented)
}

// Seek attempts to call Seek on the given File.
// If the File does not implement [ihfs.Seeker], Seek returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Seek(f ihfs.File, offset int64, whence int) (int64, error) {
	if seeker, ok := f.(ihfs.Seeker); ok {
		return seeker.Seek(offset, whence)
	}
	return 0, fmt.Errorf("seek: %w", ErrNotImplemented)
}

// Sync attempts to call Sync on the given File.
// If the File does not implement [ihfs.Syncer], Sync returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Sync(f ihfs.File) error {
	if syncer, ok := f.(ihfs.Syncer); ok {
		return syncer.Sync()
	}
	return fmt.Errorf("sync: %w", ErrNotImplemented)
}

// Truncate attempts to call Truncate on the given File.
// If the File does not implement [ihfs.Truncater], Truncate returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Truncate(f ihfs.File, size int64) error {
	if truncater, ok := f.(ihfs.Truncater); ok {
		return truncater.Truncate(size)
	}
	return fmt.Errorf("truncate: %w", ErrNotImplemented)
}

// Write attempts to call Write on the given File.
// If the File does not implement [ihfs.Writer], Write returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func Write(f ihfs.File, p []byte) (int, error) {
	if writer, ok := f.(ihfs.Writer); ok {
		return writer.Write(p)
	}
	return 0, fmt.Errorf("write: %w", ErrNotImplemented)
}

// WriteAt attempts to call WriteAt on the given File.
// If the File does not implement [ihfs.WriterAt], WriteAt returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func WriteAt(f ihfs.File, p []byte, off int64) (int, error) {
	if writerAt, ok := f.(ihfs.WriterAt); ok {
		return writerAt.WriteAt(p, off)
	}
	return 0, fmt.Errorf("write at: %w", ErrNotImplemented)
}

// WriteString attempts to call WriteString on the given File.
// If the File does not implement [ihfs.StringWriter], WriteString returns
// an error that can be checked with [errors.Is] for [ErrNotImplemented].
func WriteString(f ihfs.File, s string) (int, error) {
	if stringWriter, ok := f.(ihfs.StringWriter); ok {
		return stringWriter.WriteString(s)
	}
	return 0, fmt.Errorf("write string: %w", ErrNotImplemented)
}

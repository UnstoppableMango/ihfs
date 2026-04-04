// Package errfs provides a filesystem implementation that returns a fixed error for all operations.
// Intended for testing error handling paths.
package errfs

import (
	"time"

	"github.com/unstoppablemango/ihfs"
)

// Fs is a filesystem that returns a fixed error for all operations.
type Fs struct {
	err error
}

// New creates a new [Fs] that returns err for all operations.
func New(err error) *Fs {
	return &Fs{err: err}
}

// Open implements [ihfs.FS].
func (f *Fs) Open(_ string) (ihfs.File, error) {
	return nil, f.err
}

// Stat implements [ihfs.StatFS].
func (f *Fs) Stat(_ string) (ihfs.FileInfo, error) {
	return nil, f.err
}

// Create implements [ihfs.CreateFS].
func (f *Fs) Create(_ string) (ihfs.Writer, error) {
	return nil, f.err
}

// CreateTemp implements [ihfs.CreateTempFS].
func (f *Fs) CreateTemp(_, _ string) (ihfs.File, error) {
	return nil, f.err
}

// WriteFile implements [ihfs.WriteFileFS].
func (f *Fs) WriteFile(_ string, _ []byte, _ ihfs.FileMode) error {
	return f.err
}

// ReadFile implements [ihfs.ReadFileFS].
func (f *Fs) ReadFile(_ string) ([]byte, error) {
	return nil, f.err
}

// Chmod implements [ihfs.ChmodFS].
func (f *Fs) Chmod(_ string, _ ihfs.FileMode) error {
	return f.err
}

// Chown implements [ihfs.ChownFS].
func (f *Fs) Chown(_ string, _, _ int) error {
	return f.err
}

// Chtimes implements [ihfs.ChtimesFS].
func (f *Fs) Chtimes(_ string, _, _ time.Time) error {
	return f.err
}

// Copy implements [ihfs.CopyFS].
func (f *Fs) Copy(_ string, _ ihfs.FS) error {
	return f.err
}

// Glob implements [ihfs.GlobFS].
func (f *Fs) Glob(_ string) ([]string, error) {
	return nil, f.err
}

// Lstat implements a Lstat operation.
func (f *Fs) Lstat(_ string) (ihfs.FileInfo, error) {
	return nil, f.err
}

// Mkdir implements [ihfs.MkdirFS].
func (f *Fs) Mkdir(_ string, _ ihfs.FileMode) error {
	return f.err
}

// MkdirAll implements [ihfs.MkdirAllFS].
func (f *Fs) MkdirAll(_ string, _ ihfs.FileMode) error {
	return f.err
}

// MkdirTemp implements [ihfs.MkdirTempFS].
func (f *Fs) MkdirTemp(_, _ string) (string, error) {
	return "", f.err
}

// OpenFile implements [ihfs.OpenFileFS].
func (f *Fs) OpenFile(_ string, _ int, _ ihfs.FileMode) (ihfs.File, error) {
	return nil, f.err
}

// ReadDir implements [ihfs.ReadDirFS].
func (f *Fs) ReadDir(_ string) ([]ihfs.DirEntry, error) {
	return nil, f.err
}

// ReadDirNames implements [ihfs.ReadDirNamesFS].
func (f *Fs) ReadDirNames(_ string) ([]string, error) {
	return nil, f.err
}

// ReadLink implements [ihfs.ReadLinkFS].
func (f *Fs) ReadLink(_ string) (string, error) {
	return "", f.err
}

// Remove implements [ihfs.RemoveFS].
func (f *Fs) Remove(_ string) error {
	return f.err
}

// RemoveAll implements [ihfs.RemoveAllFS].
func (f *Fs) RemoveAll(_ string) error {
	return f.err
}

// Rename implements [ihfs.RenameFS].
func (f *Fs) Rename(_, _ string) error {
	return f.err
}

// Sub implements [ihfs.SubFS].
func (f *Fs) Sub(_ string) (ihfs.FS, error) {
	return nil, f.err
}

// Symlink implements [ihfs.SymlinkFS].
func (f *Fs) Symlink(_, _ string) error {
	return f.err
}

// TempFile implements [ihfs.TempFileFS].
func (f *Fs) TempFile(_, _ string) (string, error) {
	return "", f.err
}

package testfs

import (
	"time"

	"github.com/unstoppablemango/ihfs"
)

// Option configures a [Fs] test filesystem.
type Option func(*Fs)

// WithOpen sets the Open function on the test filesystem.
func WithOpen(fn func(string) (ihfs.File, error)) Option {
	return func(fs *Fs) {
		fs.WithOpen(fn)
	}
}

// WithStat sets the Stat function on the test filesystem.
func WithStat(fn func(string) (ihfs.FileInfo, error)) Option {
	return func(fs *Fs) {
		fs.WithStat(fn)
	}
}

// WithCreate sets the Create function on the test filesystem.
func WithCreate(fn func(string) (ihfs.File, error)) Option {
	return func(fs *Fs) {
		fs.WithCreate(fn)
	}
}

// WithWriteFile sets the WriteFile function on the test filesystem.
func WithWriteFile(fn func(string, []byte, ihfs.FileMode) error) Option {
	return func(fs *Fs) {
		fs.WithWriteFile(fn)
	}
}

// WithChmod sets the Chmod function on the test filesystem.
func WithChmod(fn func(string, ihfs.FileMode) error) Option {
	return func(fs *Fs) {
		fs.WithChmod(fn)
	}
}

// WithChown sets the Chown function on the test filesystem.
func WithChown(fn func(string, int, int) error) Option {
	return func(fs *Fs) {
		fs.WithChown(fn)
	}
}

// WithChtimes sets the Chtimes function on the test filesystem.
func WithChtimes(fn func(string, time.Time, time.Time) error) Option {
	return func(fs *Fs) {
		fs.WithChtimes(fn)
	}
}

// WithCopy sets the Copy function on the test filesystem.
func WithCopy(fn func(string, ihfs.FS) error) Option {
	return func(fs *Fs) {
		fs.WithCopy(fn)
	}
}

// WithMkdir sets the Mkdir function on the test filesystem.
func WithMkdir(fn func(string, ihfs.FileMode) error) Option {
	return func(fs *Fs) {
		fs.WithMkdir(fn)
	}
}

// WithMkdirAll sets the MkdirAll function on the test filesystem.
func WithMkdirAll(fn func(string, ihfs.FileMode) error) Option {
	return func(fs *Fs) {
		fs.WithMkdirAll(fn)
	}
}

// WithMkdirTemp sets the MkdirTemp function on the test filesystem.
func WithMkdirTemp(fn func(string, string) (string, error)) Option {
	return func(fs *Fs) {
		fs.WithMkdirTemp(fn)
	}
}

// WithRemove sets the Remove function on the test filesystem.
func WithRemove(fn func(string) error) Option {
	return func(fs *Fs) {
		fs.WithRemove(fn)
	}
}

// WithRemoveAll sets the RemoveAll function on the test filesystem.
func WithRemoveAll(fn func(string) error) Option {
	return func(fs *Fs) {
		fs.WithRemoveAll(fn)
	}
}

// WithReadDir sets the ReadDir function on the test filesystem.
func WithReadDir(fn func(string) ([]ihfs.DirEntry, error)) Option {
	return func(fs *Fs) {
		fs.WithReadDir(fn)
	}
}

// WithCreateTemp sets the CreateTemp function on the test filesystem.
func WithCreateTemp(fn func(string, string) (ihfs.File, error)) Option {
	return func(fs *Fs) {
		fs.WithCreateTemp(fn)
	}
}

// WithGlob sets the Glob function on the test filesystem.
func WithGlob(fn func(string) ([]string, error)) Option {
	return func(fs *Fs) {
		fs.WithGlob(fn)
	}
}

// WithLstat sets the Lstat function on the test filesystem.
func WithLstat(fn func(string) (ihfs.FileInfo, error)) Option {
	return func(fs *Fs) {
		fs.WithLstat(fn)
	}
}

// WithOpenFile sets the OpenFile function on the test filesystem.
func WithOpenFile(fn func(string, int, ihfs.FileMode) (ihfs.File, error)) Option {
	return func(fs *Fs) {
		fs.WithOpenFile(fn)
	}
}

// WithReadDirNames sets the ReadDirNames function on the test filesystem.
func WithReadDirNames(fn func(string) ([]string, error)) Option {
	return func(fs *Fs) {
		fs.WithReadDirNames(fn)
	}
}

// WithReadFile sets the ReadFile function on the test filesystem.
func WithReadFile(fn func(string) ([]byte, error)) Option {
	return func(fs *Fs) {
		fs.WithReadFile(fn)
	}
}

// WithReadLink sets the ReadLink function on the test filesystem.
func WithReadLink(fn func(string) (string, error)) Option {
	return func(fs *Fs) {
		fs.WithReadLink(fn)
	}
}

// WithRename sets the Rename function on the test filesystem.
func WithRename(fn func(string, string) error) Option {
	return func(fs *Fs) {
		fs.WithRename(fn)
	}
}

// WithSub sets the Sub function on the test filesystem.
func WithSub(fn func(string) (ihfs.FS, error)) Option {
	return func(fs *Fs) {
		fs.WithSub(fn)
	}
}

// WithSymlink sets the Symlink function on the test filesystem.
func WithSymlink(fn func(string, string) error) Option {
	return func(fs *Fs) {
		fs.WithSymlink(fn)
	}
}

// WithTempFile sets the TempFile function on the test filesystem.
func WithTempFile(fn func(string, string) (string, error)) Option {
	return func(fs *Fs) {
		fs.WithTempFile(fn)
	}
}

// Package factory provides a factory-pattern test filesystem for use in unit tests.
package factory

import (
	"errors"

	"github.com/unstoppablemango/ihfs"
)

type (
	// OpenFunc is a function that opens a file by name.
	OpenFunc func(string) (ihfs.File, error)
	// StatFunc is a function that returns file info for the named file.
	StatFunc func(string) (ihfs.FileInfo, error)
)

// ErrNotMocked is returned when an operation has no mock configured.
var ErrNotMocked = errors.New("operation has no mock")

// Fs is a test filesystem that returns pre-configured responses in sequence.
type Fs struct {
	name string
	open []OpenFunc
	stat []StatFunc
}

// NewFs creates a new factory [Fs].
func NewFs() *Fs {
	return &Fs{name: "testfs/factory"}
}

// Named sets the name of the filesystem and returns it.
func (f *Fs) Named(name string) *Fs {
	f.name = name
	return f
}

// Name returns the filesystem name.
func (f *Fs) Name() string {
	return f.name
}

// WithOpen appends Open functions to the factory queue.
func (f *Fs) WithOpen(open ...OpenFunc) *Fs {
	f.open = append(f.open, open...)
	return f
}

// SetOpen replaces the Open function queue.
func (f *Fs) SetOpen(open ...OpenFunc) *Fs {
	f.open = open
	return f
}

// Open implements [ihfs.FS] by consuming the next queued Open function.
func (f *Fs) Open(path string) (ihfs.File, error) {
	if len(f.open) == 0 {
		return nil, ErrNoMocks
	}

	open := f.open[0]
	f.open = f.open[1:]
	return open(path)
}

// WithStat appends Stat functions to the factory queue.
func (f *Fs) WithStat(stat ...StatFunc) *Fs {
	f.stat = append(f.stat, stat...)
	return f
}

// SetStat replaces the Stat function queue.
func (f *Fs) SetStat(stat ...StatFunc) *Fs {
	f.stat = stat
	return f
}

// Stat implements [ihfs.StatFS] by consuming the next queued Stat function.
func (f *Fs) Stat(path string) (ihfs.FileInfo, error) {
	if len(f.stat) == 0 {
		return nil, ErrNoMocks
	}

	stat := f.stat[0]
	f.stat = f.stat[1:]
	return stat(path)
}

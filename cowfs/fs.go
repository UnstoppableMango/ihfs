package cowfs

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/unmango/go/fopt"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/fsutil/try"
)

// Fs implements a copy-on-write filesystem. Changes to the file system will
// only be made in the overlay. Changing an existing file in the base layer
// which is not present in the overlay will copy the file to the overlay.
//
// The implementation is based heavily on [afero.CopyOnWriteFs].
type Fs struct {
	base  ihfs.FS
	layer ihfs.FS
	merge MergeStrategy
}

// New creates a new copy-on-write filesystem with the given base and layer.
func New(base, layer ihfs.FS, options ...Option) *Fs {
	f := &Fs{
		base:  base,
		layer: layer,
		merge: DefaultMergeStrategy,
	}
	fopt.ApplyAll(f, options)

	return f
}

// Open implements [fs.FS].
func (f *Fs) Open(name string) (ihfs.File, error) {
	if inBase, err := f.isInBase(name); err != nil {
		return nil, err
	} else if inBase {
		return f.base.Open(name)
	}

	if isDir, err := try.IsDir(f.layer, name); err != nil {
		return nil, err
	} else if !isDir {
		return f.layer.Open(name)
	}

	if isDir, err := try.IsDir(f.base, name); !isDir || err != nil {
		return f.layer.Open(name)
	}

	bFile, bErr := f.base.Open(name)
	lFile, lErr := f.layer.Open(name)

	if bErr == nil && lErr == nil {
		return newFile(bFile, lFile, f.merge), nil
	}

	// TODO: possible file handle leaking
	// https://github.com/UnstoppableMango/ihfs/pull/14#discussion_r2737484821

	return nil, &ihfs.PathError{
		Op:   "open",
		Path: name,
		Err: errors.Join(
			fmt.Errorf("base: %w", bErr),
			fmt.Errorf("layer: %w", lErr),
		),
	}
}

func (f *Fs) isInBase(path string) (bool, error) {
	if exists, err := try.Exists(f.layer, path); err != nil {
		return false, fmt.Errorf("layer: %w", err)
	} else if exists {
		return false, nil
	}

	if _, err := try.Stat(f.base, path); err != nil {
		switch {
		case errors.Is(err, ihfs.ErrNotExist):
			fallthrough
		case errors.Is(err, syscall.ENOTDIR):
			return false, nil
		default:
			return false, fmt.Errorf("base: %w", err)
		}
	}

	return true, nil
}

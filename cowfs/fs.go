package cowfs

import (
	"errors"
	"fmt"
	"syscall"

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
}

// New creates a new copy-on-write filesystem with the given base and layer.
func New(base, layer ihfs.FS) *Fs {
	return &Fs{
		base:  base,
		layer: layer,
	}
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

	if bErr != nil || lErr != nil {
		return nil, &ihfs.PathError{
			Op:   "open",
			Path: name,
			Err: errors.Join(
				fmt.Errorf("base: %w", bErr),
				fmt.Errorf("layer: %w", lErr),
			),
		}
	}

	return &File{
		name:  name,
		base:  bFile,
		layer: lFile,
	}, nil
}

func (f *Fs) isInBase(path string) (bool, error) {
	if exists, _ := try.Exists(f.layer, path); exists {
		return false, nil
	}

	_, err := try.Stat(f.base, path)
	if err != nil {
		if errors.Is(err, ihfs.ErrNotExist) {
			return false, nil
		}
		if errors.Is(err, syscall.ENOENT) {
			return false, nil
		}
		if errors.Is(err, syscall.ENOTDIR) {
			return false, nil
		}
	}

	return true, err
}

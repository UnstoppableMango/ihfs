package cowfs

import (
	"errors"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/fsutil/try"
)

type Fs struct {
	base  ihfs.FS
	layer ihfs.FS
}

func (fs *Fs) Open(path string) (ihfs.File, error) {
	if inBase, _ := fs.isInBase(path); inBase {
		return fs.base.Open(path)
	}
	return fs.layer.Open(path)
}

func (fs *Fs) isInBase(path string) (bool, error) {
	if exists, _ := try.Exists(fs.layer, path); exists {
		return false, nil
	}

	_, err := try.Stat(fs.base, path)
	if err != nil {
		if errors.Is(err, ihfs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

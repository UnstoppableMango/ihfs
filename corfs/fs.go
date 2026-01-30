package corfs

import (
	"errors"
	"fmt"
	"io"
	"syscall"
	"time"

	"github.com/unmango/go/fopt"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/fsutil/try"
	"github.com/unstoppablemango/ihfs/union"
)

// Fs implements a cache-on-read filesystem. When files are read from
// the base filesystem, they are cached in the layer. Future reads come from the
// cached version until it expires or is invalidated.
//
// If the cache duration is 0, cache time will be unlimited, i.e. once
// a file is in the layer, the base will never be read again for this file.
//
// For cache times greater than 0, the modification time of a file is
// checked. Note that a lot of file system implementations only allow a
// resolution of a second for timestamps.
//
// The implementation is based heavily on [afero.CacheOnReadFs].
type Fs struct {
	base      ihfs.FS
	layer     ihfs.FS
	cacheTime time.Duration
}

// New creates a new cache-on-read filesystem with the given base and layer.
// The cacheTime parameter determines how long cached files are valid.
// If cacheTime is 0, files are cached indefinitely.
func New(base, layer ihfs.FS, options ...Option) *Fs {
	f := &Fs{
		base:      base,
		layer:     layer,
		cacheTime: 0,
	}
	fopt.ApplyAll(f, options)

	return f
}

type cacheState int

const (
	// cacheMiss: not present in the overlay, unknown if it exists in the base
	cacheMiss cacheState = iota
	// cacheStale: present in the overlay and in base, base file is newer
	cacheStale
	// cacheHit: present in the overlay - with cache time == 0 it may exist in the base,
	// with cacheTime > 0 it exists in the base and is same age or newer in the overlay
	cacheHit
	// cacheLocal: happens if someone writes directly to the overlay without
	// going through this union
	cacheLocal
)

// cacheStatus checks the cache status of a file
func (f *Fs) cacheStatus(name string) (state cacheState, fi ihfs.FileInfo, err error) {
	var lfi, bfi ihfs.FileInfo
	lfi, err = try.Stat(f.layer, name)
	if err == nil {
		if f.cacheTime == 0 {
			return cacheHit, lfi, nil
		}
		if lfi.ModTime().Add(f.cacheTime).Before(time.Now()) {
			bfi, err = try.Stat(f.base, name)
			if err != nil {
				return cacheLocal, lfi, nil
			}
			if bfi.ModTime().After(lfi.ModTime()) {
				return cacheStale, bfi, nil
			}
		}
		return cacheHit, lfi, nil
	}

	if errors.Is(err, ihfs.ErrNotExist) || errors.Is(err, syscall.ENOENT) {
		return cacheMiss, nil, nil
	}

	return cacheMiss, nil, err
}

// copyToLayer copies a file from the base to the layer
func (f *Fs) copyToLayer(name string) error {
	// Open the file in the base filesystem
	bFile, err := f.base.Open(name)
	if err != nil {
		return fmt.Errorf("failed to open base file: %w", err)
	}
	defer bFile.Close()

	// Get the file info to check if it's a directory
	bInfo, err := bFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat base file: %w", err)
	}

	// If it's a directory, create it in the layer
	if bInfo.IsDir() {
		if mkdirer, ok := f.layer.(ihfs.MkdirAllFS); ok {
			return mkdirer.MkdirAll(name, bInfo.Mode())
		}
		return fmt.Errorf("layer filesystem does not support MkdirAll")
	}

	// Ensure parent directories exist in the layer
	if mkdirer, ok := f.layer.(ihfs.MkdirAllFS); ok {
		parent := name
		// Find the parent directory
		for i := len(name) - 1; i >= 0; i-- {
			if name[i] == '/' {
				parent = name[:i]
				break
			}
		}
		if parent != name && parent != "" {
			if err := mkdirer.MkdirAll(parent, 0755); err != nil {
				return fmt.Errorf("failed to create parent directories: %w", err)
			}
		}
	}

	// Create the file in the layer
	var lFile ihfs.File
	if creator, ok := f.layer.(ihfs.CreateFS); ok {
		lFile, err = creator.Create(name)
		if err != nil {
			return fmt.Errorf("failed to create layer file: %w", err)
		}
		defer func() {
			if closeErr := lFile.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
		}()
	} else {
		return fmt.Errorf("layer filesystem does not support Create")
	}

	// Copy the contents - lFile needs to be a Writer
	if writer, ok := lFile.(ihfs.Writer); ok {
		if _, err := io.Copy(writer, bFile); err != nil {
			// Clean up the partially created file
			if remover, ok := f.layer.(ihfs.RemoveFS); ok {
				_ = remover.Remove(name)
			}
			return fmt.Errorf("failed to copy file contents: %w", err)
		}
	} else {
		// Clean up the created file since we can't write to it
		if remover, ok := f.layer.(ihfs.RemoveFS); ok {
			_ = remover.Remove(name)
		}
		return fmt.Errorf("layer file does not support Write")
	}

	// Copy the modification time
	if chtimer, ok := f.layer.(ihfs.ChtimesFS); ok {
		if err := chtimer.Chtimes(name, bInfo.ModTime(), bInfo.ModTime()); err != nil {
			// Log but don't fail - the file was copied successfully
			return nil
		}
	}

	return nil
}

// Open implements [fs.FS].
func (f *Fs) Open(name string) (ihfs.File, error) {
	st, fi, err := f.cacheStatus(name)
	if err != nil {
		return nil, err
	}

	switch st {
	case cacheLocal:
		return f.layer.Open(name)

	case cacheMiss:
		bfi, err := try.Stat(f.base, name)
		if err != nil {
			return nil, err
		}
		if bfi.IsDir() {
			// For directories, fall through to merge logic below
		} else {
			if err := f.copyToLayer(name); err != nil {
				return nil, err
			}
			return f.layer.Open(name)
		}
	case cacheStale:
		if !fi.IsDir() {
			if err := f.copyToLayer(name); err != nil {
				return nil, err
			}
			return f.layer.Open(name)
		}
	case cacheHit:
		if !fi.IsDir() {
			return f.layer.Open(name)
		}
	}

	// the dirs from cacheHit, cacheStale, and cacheMiss fall down here:
	bfile, bErr := f.base.Open(name)
	lfile, lErr := f.layer.Open(name)

	// Only ignore base errors if it's a not-exist error
	if bErr != nil && !errors.Is(bErr, ihfs.ErrNotExist) && lfile == nil {
		return nil, bErr
	}

	if lErr != nil && bfile == nil {
		return nil, lErr
	}
	return union.NewFile(bfile, lfile), nil
}

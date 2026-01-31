package union

import (
	"io"
	"path/filepath"
	"syscall"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/fsutil/try"
)

// CopyToLayer copies a file from the base filesystem to the layer filesystem.
// This is typically used in copy-on-write filesystems when a file needs to be
// modified - it first copies the file from the base (read-only) layer to the
// layer (writable) layer, then modifications can be made.
//
// The function:
//   - Creates any necessary parent directories in the layer
//   - Copies the file content from base to layer
//   - Preserves the file's modification time
//
// Returns an error if:
//   - The file cannot be opened in the base
//   - Parent directories cannot be created in the layer
//   - The file cannot be created in the layer
//   - The copy operation fails
//   - File metadata cannot be retrieved or set
func CopyToLayer(base, layer ihfs.FS, name string) error {
	file, err := base.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	return copyFile(layer, name, file)
}

// copyFile is an internal helper that performs the actual file copy operation.
// It takes an already-opened file handle from the base filesystem and copies
// it to the layer filesystem, preserving metadata.
func copyFile(layer ihfs.FS, name string, file ihfs.File) error {
	// First make sure the directory exists
	dir := filepath.Dir(name)
	if exists, err := try.Exists(layer, dir); err != nil {
		return err
	} else if !exists {
		if err := try.MkdirAll(layer, dir, 0o777); err != nil {
			return err
		}
	}

	// Create the file on the overlay
	lfh, err := try.Create(layer, name)
	if err != nil {
		return err
	}

	// Ensure the file supports writing
	writer, ok := lfh.(io.Writer)
	if !ok {
		lfh.Close()
		try.Remove(layer, name)
		return &ihfs.PathError{
			Op:   "copy",
			Path: name,
			Err:  syscall.ENOTSUP,
		}
	}

	// Copy the content
	n, err := io.Copy(writer, file)
	if err != nil {
		// If anything fails, clean up the file
		lfh.Close()
		try.Remove(layer, name)
		return err
	}

	// Verify the copy was complete
	bfi, err := file.Stat()
	if err != nil {
		lfh.Close()
		try.Remove(layer, name)
		return err
	}

	if bfi.Size() != n {
		lfh.Close()
		try.Remove(layer, name)
		return syscall.EIO
	}

	// Close the file before setting times
	err = lfh.Close()
	if err != nil {
		try.Remove(layer, name)
		return err
	}

	// Preserve modification time
	return try.Chtimes(layer, name, bfi.ModTime(), bfi.ModTime())
}

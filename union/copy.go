package union

import (
	"io"
	"path/filepath"
	"syscall"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/try"
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
	// TODO: Check for ihfs.CopyFS interface and use that if available
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
	lFile, err := try.Create(layer, name)
	if err != nil {
		return err
	}

	// Ensure the file supports writing
	writer, ok := lFile.(io.Writer)
	if !ok {
		lFile.Close()
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
		lFile.Close()
		try.Remove(layer, name)
		return err
	}

	// Verify the copy was complete
	bFile, err := file.Stat()
	if err != nil {
		lFile.Close()
		try.Remove(layer, name)
		return err
	}

	if bFile.Size() != n {
		lFile.Close()
		try.Remove(layer, name)
		return syscall.EIO
	}

	// Close the file before setting times
	if err = lFile.Close(); err != nil {
		try.Remove(layer, name)
		return err
	}

	// Preserve modification time
	return try.Chtimes(layer, name, bFile.ModTime(), bFile.ModTime())
}

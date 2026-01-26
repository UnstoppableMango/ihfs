// Package cowfs implements a copy-on-write filesystem. Changes to the file system will
// only be made in the overlay. Changing an existing file in the base layer
// which is not present in the overlay will copy the file to the overlay.
//
// The implementation is based heavily on [afero.CopyOnWriteFs].
package cowfs

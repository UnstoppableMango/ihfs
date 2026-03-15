package ctrfs

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/unstoppablemango/ihfs/tarfs"
)

// ImageFS wraps a [v1.Image] as a read-only file system.
// The image layers are merged with whiteout entries applied via [mutate.Extract],
// presenting the final filesystem view of the image.
//
// Call [ImageFS.Close] when the FS is no longer needed to release the underlying stream.
type ImageFS struct {
	*tarfs.TarFile
}

// FromImage creates a read-only [io/fs.FS] from a [v1.Image].
// The returned FS presents the merged, whiteout-resolved view of all image layers.
func FromImage(img v1.Image) *ImageFS {
	return &ImageFS{tarfs.FromReader("", mutate.Extract(img))}
}

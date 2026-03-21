package image

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/ctrfs/v1/layer"
	"github.com/unstoppablemango/ihfs/tarfs"
)

// FS wraps a [v1.Image] as a read-only file system.
// The image layers are merged with whiteout entries applied via [mutate.Extract],
// presenting the final filesystem view of the image.
//
// Call [FS.Close] when the FS is no longer needed to release the underlying stream.
type FS struct {
	*tarfs.TarFile
}

// From creates a read-only [io/fs.FS] from a [v1.Image].
// The returned FS presents the merged, whiteout-resolved view of all image layers.
func From(img v1.Image) *FS {
	return &FS{tarfs.FromReader("", mutate.Extract(img))}
}

// Create appends a new layer built from fsys onto base and returns the resulting image.
func Create(base v1.Image, fsys ihfs.FS, dir string) (v1.Image, error) {
	l, err := layer.Create(fsys, dir)
	if err != nil {
		return nil, err
	}
	return mutate.AppendLayers(base, l)
}

package v1

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/tarfs"
)

// FromImage creates a read-only [io/fs.FS] from a [v1.Image].
// The returned FS presents the merged, whiteout-resolved view of all image layers.
func FromImage(img v1.Image) *FS {
	rc := mutate.Extract(img)
	return &FS{Fs: tarfs.FromReader(rc), closer: rc}
}

// AppendFS appends a new layer built from fsys onto base and returns the resulting image.
func AppendFS(base v1.Image, fsys ihfs.FS, dir string) (v1.Image, error) {
	l, err := ToLayer(fsys, dir)
	if err != nil {
		return nil, err
	}
	return mutate.AppendLayers(base, l)
}

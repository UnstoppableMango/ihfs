// Package ctrfs provides read-only fs.FS implementations for OCI container images and layers,
// and write helpers for producing new OCI layers from an fs.FS.
//
// # Reading
//
// [FromImage] wraps a [github.com/google/go-containerregistry/pkg/v1.Image] as a read-only
// [io/fs.FS]. All image layers are merged with whiteout entries applied so the returned FS
// presents the final filesystem view of the image.
//
//	img, _ := remote.Image(ref)
//	fsys := ctrfs.FromImage(img)
//	defer fsys.Close()
//	data, _ := fs.ReadFile(fsys, "etc/os-release")
//
// [FromLayer] wraps a single [github.com/google/go-containerregistry/pkg/v1.Layer] as a
// read-only [io/fs.FS].
//
//	layers, _ := img.Layers()
//	fsys, _ := ctrfs.FromLayer(layers[0])
//	defer fsys.Close()
//
// # Writing
//
// [ToLayer] walks an [io/fs.FS] and produces a new [github.com/google/go-containerregistry/pkg/v1.Layer]
// containing those files as a compressed tar archive.
//
// [ToImage] appends a new layer from an [io/fs.FS] onto a base image and returns the result.
//
//	layer, _ := ctrfs.ToLayer(myFS, ".")
//	newImg, _ := ctrfs.ToImage(baseImg, myFS, ".")
package ctrfs

package ctrfs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/tarfs"
)

// LayerFS wraps a [v1.Layer] as a read-only file system.
//
// Call [LayerFS.Close] when the FS is no longer needed to release the underlying stream.
type LayerFS struct {
	*tarfs.TarFile
}

// FromLayer creates a read-only [io/fs.FS] from a [v1.Layer].
func FromLayer(layer v1.Layer) (*LayerFS, error) {
	rc, err := layer.Uncompressed()
	if err != nil {
		return nil, err
	}
	return &LayerFS{tarfs.FromReader("", rc)}, nil
}

// ToLayer creates a [v1.Layer] from the files in fsys rooted at dir.
// The resulting layer contains all files as a gzip-compressed tar archive.
func ToLayer(fsys ihfs.FS, dir string) (v1.Layer, error) {
	var compressed bytes.Buffer
	if err := writeLayer(fsys, dir, &compressed); err != nil {
		return nil, err
	}
	return tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(compressed.Bytes())), nil
	})
}

// writeLayer writes a gzip-compressed tar archive of fsys rooted at dir to w.
func writeLayer(fsys ihfs.FS, dir string, w io.Writer) error {
	gw := gzip.NewWriter(w)
	tw := tar.NewWriter(gw)

	err := fs.WalkDir(fsys, dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := entryName(p, dir)

		info, err := d.Info()
		if err != nil {
			return err
		}

		var link string
		if d.Type()&fs.ModeSymlink != 0 {
			if link, err = fs.ReadLink(fsys, p); err != nil {
				return err
			}
		}

		hdr, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		hdr.Name = name
		if d.IsDir() && name != "." {
			hdr.Name += "/"
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if d.Type().IsRegular() {
			f, err := fsys.Open(p)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gw.Close()
}

// ToImage appends a new layer built from fsys onto base and returns the resulting image.
func ToImage(base v1.Image, fsys ihfs.FS, dir string) (v1.Image, error) {
	layer, err := ToLayer(fsys, dir)
	if err != nil {
		return nil, err
	}
	return mutate.AppendLayers(base, layer)
}

// entryName computes the tar entry name for a path p relative to dir.
func entryName(p, dir string) string {
	if p == dir {
		return "."
	}
	if dir == "." {
		return p
	}
	return strings.TrimPrefix(p, dir+"/")
}

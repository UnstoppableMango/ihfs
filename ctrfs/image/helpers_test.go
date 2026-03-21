package image_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// tarEntry holds a tar header and optional file content for test layer construction.
type tarEntry struct {
	hdr  *tar.Header
	data string
}

// makeLayer creates a v1.Layer from a slice of tar entries.
// The opener produces a gzip-compressed tar archive as expected by go-containerregistry.
func makeLayer(entries []tarEntry) (v1.Layer, error) {
	return tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gw)
		for _, e := range entries {
			if err := tw.WriteHeader(e.hdr); err != nil {
				return nil, err
			}
			if e.data != "" {
				if _, err := tw.Write([]byte(e.data)); err != nil {
					return nil, err
				}
			}
		}
		if err := tw.Close(); err != nil {
			return nil, err
		}
		if err := gw.Close(); err != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
	})
}

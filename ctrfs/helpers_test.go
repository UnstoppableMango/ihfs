package ctrfs_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/testfs"
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

// tarNames reads all entry names from a tar stream, returning an error if the stream is malformed.
func tarNames(r io.Reader) ([]string, error) {
	tr := tar.NewReader(r)
	var names []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name != "." {
			names = append(names, hdr.Name)
		}
	}
	return names, nil
}

func rootDirStat(name string) (ihfs.FileInfo, error) {
	fi := testfs.NewFileInfo(name)
	fi.IsDirFunc = func() bool { return name == "." }
	fi.ModeFunc = func() fs.FileMode {
		if name == "." {
			return fs.ModeDir
		}
		return 0
	}
	return fi, nil
}

// errLayer is a v1.Layer whose methods all return a fixed error.
type errLayer struct {
	err error
}

func (e *errLayer) Digest() (v1.Hash, error)             { return v1.Hash{}, e.err }
func (e *errLayer) DiffID() (v1.Hash, error)             { return v1.Hash{}, e.err }
func (e *errLayer) Compressed() (io.ReadCloser, error)   { return nil, e.err }
func (e *errLayer) Uncompressed() (io.ReadCloser, error) { return nil, e.err }
func (e *errLayer) Size() (int64, error)                 { return 0, e.err }
func (e *errLayer) MediaType() (types.MediaType, error)  { return "", e.err }

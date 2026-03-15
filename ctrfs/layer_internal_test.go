package ctrfs

import (
	"errors"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

// errWriter is an io.Writer that always returns an error.
type errWriter struct {
	err error
}

func (e *errWriter) Write(_ []byte) (int, error) {
	return 0, e.err
}

func rootDirStatInternal(name string) (ihfs.FileInfo, error) {
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

var _ = Describe("writeLayer", func() {
	It("should propagate WriteHeader errors", func() {
		writeErr := errors.New("write error")
		fsys := testfs.New(
			testfs.WithStat(rootDirStatInternal),
			testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
				return nil, nil
			}),
		)

		err := writeLayer(fsys, ".", &errWriter{err: writeErr})

		Expect(err).To(HaveOccurred())
	})
})

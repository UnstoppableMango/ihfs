package errfs_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/errfs"
)

var _ = Describe("File", func() {
	var (
		sentinel = errors.New("sentinel error")
		file     *errfs.File
	)

	BeforeEach(func() {
		file = errfs.NewFile(sentinel)
	})

	It("should return the error from Close", func() {
		err := file.Close()
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Read", func() {
		_, err := file.Read(make([]byte, 10))
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Stat", func() {
		_, err := file.Stat()
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from ReadDir", func() {
		_, err := file.ReadDir(0)
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from ReadDirNames", func() {
		_, err := file.ReadDirNames(0)
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Seek", func() {
		_, err := file.Seek(0, 0)
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Write", func() {
		_, err := file.Write([]byte("data"))
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from ReadAt", func() {
		_, err := file.ReadAt(make([]byte, 10), 0)
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from WriteAt", func() {
		_, err := file.WriteAt([]byte("data"), 0)
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from WriteString", func() {
		_, err := file.WriteString("data")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Sync", func() {
		err := file.Sync()
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Truncate", func() {
		err := file.Truncate(0)
		Expect(err).To(MatchError(sentinel))
	})

	It("should implement ihfs.File", func() {
		var _ ihfs.File = file
	})

	It("should implement ihfs.ReadDirFile", func() {
		var _ ihfs.ReadDirFile = file
	})

	It("should implement ihfs.Writer", func() {
		var _ ihfs.Writer = file
	})
})

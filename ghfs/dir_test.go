package ghfs_test

import (
	"io"
	"io/fs"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/ghfs"
)

var _ = Describe("Dir", func() {
	var dir *ghfs.File

	BeforeEach(func() {
		f, err := ghfs.New().Open(".")
		Expect(err).NotTo(HaveOccurred())
		var ok bool
		dir, ok = f.(*ghfs.File)
		Expect(ok).To(BeTrue())
	})

	It("should return error on Read", func() {
		n, err := dir.Read(make([]byte, 10))
		Expect(n).To(Equal(0))
		Expect(err).To(MatchError(fs.ErrInvalid))
	})

	It("should succeed on Close", func() {
		Expect(dir.Close()).To(Succeed())
	})

	It("should return FileInfo from Stat", func() {
		info, err := dir.Stat()
		Expect(err).NotTo(HaveOccurred())
		Expect(info).NotTo(BeNil())
	})

	It("should return io.EOF on ReadDir with n>0", func() {
		entries, err := dir.ReadDir(1)
		Expect(entries).To(BeEmpty())
		Expect(err).To(Equal(io.EOF))
	})

	It("should return empty slice on ReadDir with n<=0", func() {
		entries, err := dir.ReadDir(-1)
		Expect(entries).To(BeEmpty())
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("FileInfo", func() {
		var info fs.FileInfo

		BeforeEach(func() {
			var err error
			info, err = dir.Stat()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return '.' for Name", func() {
			Expect(info.Name()).To(Equal("."))
		})

		It("should return true for IsDir", func() {
			Expect(info.IsDir()).To(BeTrue())
		})

		It("should return a directory mode", func() {
			Expect(info.Mode()).To(Equal(fs.ModeDir | 0555))
		})

		It("should return zero time for ModTime", func() {
			Expect(info.ModTime()).To(Equal(time.Time{}))
		})

		It("should return 0 for Size", func() {
			Expect(info.Size()).To(Equal(int64(0)))
		})

		It("should return nil for Sys", func() {
			Expect(info.Sys()).To(BeNil())
		})
	})
})

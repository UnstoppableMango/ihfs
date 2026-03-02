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
	var dir *ghfs.Dir

	BeforeEach(func() {
		f, err := ghfs.New().Open(".")
		Expect(err).NotTo(HaveOccurred())
		var ok bool
		dir, ok = f.(*ghfs.Dir)
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

	It("should return itself from Stat", func() {
		info, err := dir.Stat()
		Expect(err).NotTo(HaveOccurred())
		Expect(info).To(BeIdenticalTo(dir))
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
		It("should return '.' for Name", func() {
			Expect(dir.Name()).To(Equal("."))
		})

		It("should return true for IsDir", func() {
			Expect(dir.IsDir()).To(BeTrue())
		})

		It("should return a directory mode", func() {
			Expect(dir.Mode()).To(Equal(fs.ModeDir | 0555))
		})

		It("should return zero time for ModTime", func() {
			Expect(dir.ModTime()).To(Equal(time.Time{}))
		})

		It("should return 0 for Size", func() {
			Expect(dir.Size()).To(Equal(int64(0)))
		})

		It("should return nil for Sys", func() {
			Expect(dir.Sys()).To(BeNil())
		})
	})
})

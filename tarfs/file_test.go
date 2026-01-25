package tarfs_test

import (
	"io"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/tarfs"
)

var _ = Describe("File", func() {
	var (
		tfs  *tarfs.Fs
		file fs.File
	)

	BeforeEach(func() {
		var err error
		tfs, err = tarfs.Open("../testdata/test.tar")
		Expect(err).NotTo(HaveOccurred())

		file, err = tfs.Open("tartest/test.txt")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Read", func() {
		It("should read file contents", func() {
			buf := make([]byte, 12)
			n, err := file.Read(buf)

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(12))
			Expect(string(buf)).To(Equal("test content"))
		})

		It("should return EOF when reading past end", func() {
			_, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())

			buf := make([]byte, 10)
			n, err := file.Read(buf)

			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(0))
		})

		It("should support multiple reads", func() {
			buf1 := make([]byte, 5)
			n1, err := file.Read(buf1)
			Expect(err).NotTo(HaveOccurred())
			Expect(n1).To(Equal(5))
			Expect(string(buf1)).To(Equal("test "))

			buf2 := make([]byte, 7)
			n2, err := file.Read(buf2)
			Expect(err).NotTo(HaveOccurred())
			Expect(n2).To(Equal(7))
			Expect(string(buf2)).To(Equal("content"))
		})
	})

	Describe("Stat", func() {
		It("should return file info", func() {
			info, err := file.Stat()

			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Name()).To(Equal("test.txt"))
			Expect(info.Size()).To(Equal(int64(13)))
			Expect(info.IsDir()).To(BeFalse())
		})
	})

	Describe("Close", func() {
		It("should close without error", func() {
			err := file.Close()

			Expect(err).NotTo(HaveOccurred())
		})

		It("should be idempotent", func() {
			err := file.Close()
			Expect(err).NotTo(HaveOccurred())

			err = file.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Name", func() {
		It("should return the file path", func() {
			tf, ok := file.(*tarfs.File)
			Expect(ok).To(BeTrueBecause("file is a *tarfs.File"))
			Expect(tf.Name()).To(Equal("tartest/test.txt"))
		})
	})
})

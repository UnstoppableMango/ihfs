package tarfs_test

import (
	"io"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/tarfs"
)

var _ = Describe("Fs", func() {
	Describe("Open", func() {
		It("should open a tar file", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")

			Expect(err).NotTo(HaveOccurred())
			Expect(tfs).NotTo(BeNil())
		})

		It("should return error for nonexistent file", func() {
			tfs, err := tarfs.Open("../testdata/nonexistent.tar")

			Expect(err).To(HaveOccurred())
			Expect(tfs).To(BeNil())
		})
	})

	Describe("Name", func() {
		It("should return the tar file name", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			name := tfs.Name()

			Expect(name).To(Equal("../testdata/test.tar"))
		})
	})

	Describe("Open file", func() {
		var tfs *tarfs.Fs

		BeforeEach(func() {
			var err error
			tfs, err = tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should open a file from the tar archive", func() {
			file, err := tfs.Open("tartest/test.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
		})

		It("should read file contents", func() {
			file, err := tfs.Open("tartest/test.txt")
			Expect(err).NotTo(HaveOccurred())

			content, err := io.ReadAll(file)

			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("test content\n"))
		})

		It("should return error for nonexistent file in tar", func() {
			file, err := tfs.Open("nonexistent.txt")

			Expect(err).To(MatchError(ContainSubstring("file does not exist")))
			Expect(file).To(BeNil())
		})

		It("should return cached file on subsequent opens", func() {
			file1, err := tfs.Open("tartest/test.txt")
			Expect(err).NotTo(HaveOccurred())

			file2, err := tfs.Open("tartest/test.txt")
			Expect(err).NotTo(HaveOccurred())

			Expect(file2).To(Equal(file1))
		})

		It("should open multiple files from the archive", func() {
			file1, err := tfs.Open("tartest/test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(file1).NotTo(BeNil())

			file2, err := tfs.Open("tartest/another.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(file2).NotTo(BeNil())

			content1, err := io.ReadAll(file1)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content1)).To(Equal("test content\n"))

			content2, err := io.ReadAll(file2)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content2)).To(Equal("another file\n"))
		})
	})

	Describe("OpenFS", func() {
		It("should open a tar file from a custom FS", func() {
			tfs, err := tarfs.OpenFS(os.DirFS(".."), "testdata/test.tar")

			Expect(err).NotTo(HaveOccurred())
			Expect(tfs).NotTo(BeNil())
		})

		It("should read files from tar opened with custom FS", func() {
			tfs, err := tarfs.OpenFS(os.DirFS(".."), "testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			file, err := tfs.Open("tartest/test.txt")
			Expect(err).NotTo(HaveOccurred())

			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("test content\n"))
		})
	})

	Describe("error handling", func() {
		It("should return error when tar reader fails during Open", func() {
			tfs, err := tarfs.Open("../testdata/corrupted.tar")
			Expect(err).NotTo(HaveOccurred())

			file, err := tfs.Open("any-file.txt")

			Expect(err).To(HaveOccurred())
			Expect(file).To(BeNil())
		})

		It("should return error when reading truncated tar", func() {
			tfs, err := tarfs.Open("../testdata/truncated.tar")
			Expect(err).NotTo(HaveOccurred())

			file, err := tfs.Open("any-file.txt")

			Expect(err).To(HaveOccurred())
			Expect(file).To(BeNil())
		})
	})
})

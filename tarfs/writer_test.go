package tarfs_test

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/tarfs"
)

// failAfterWriter allows n Write calls to succeed, then returns err on all subsequent calls.
type failAfterWriter struct {
	w   io.Writer
	n   int
	err error
}

func (f *failAfterWriter) Write(p []byte) (int, error) {
	if f.n == 0 {
		return 0, f.err
	}
	f.n--
	return f.w.Write(p)
}

var _ = Describe("Writer", func() {
	Describe("NewWriter", func() {
		It("should create a writer", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})

			Expect(w).NotTo(BeNil())
		})

		It("should use an existing tar.Writer directly", func() {
			tw := tar.NewWriter(&bytes.Buffer{})

			w := tarfs.NewWriter(tw)

			Expect(w).NotTo(BeNil())
		})
	})

	Describe("Close", func() {
		It("should finalize the tar archive", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			Expect(w.Close()).To(Succeed())

			tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))
			_, err := tfs.Open("nonexistent.txt")
			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("should return error when underlying writer fails on close", func() {
			writeErr := errors.New("write failed")
			fw := &failAfterWriter{w: &bytes.Buffer{}, n: 0, err: writeErr}
			w := tarfs.NewWriter(fw)

			err := w.Close()

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Open", func() {
		It("should return ErrPermission", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})

			file, err := w.Open("test.txt")

			Expect(file).To(BeNil())
			Expect(err).To(MatchError(ihfs.ErrPermission))
		})
	})

	Describe("Create", func() {
		It("should return a writable file handle for a valid path", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			file, err := w.Create("test.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			DeferCleanup(file.Close)
		})

		It("should return ErrInvalid for an invalid path", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})

			file, err := w.Create("../invalid")

			Expect(file).To(BeNil())
			Expect(err).To(MatchError(ihfs.ErrInvalid))
		})

		It("should write buffered content as a tar entry on Close", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			file, err := w.Create("hello.txt")
			Expect(err).NotTo(HaveOccurred())

			_, err = file.(io.Writer).Write([]byte("hello world"))
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Close()).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))
			f, err := tfs.Open("hello.txt")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(f.Close)
			content, err := io.ReadAll(f)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("hello world"))
		})
	})

	Describe("writerFile", func() {
		var (
			w   *tarfs.Writer
			buf bytes.Buffer
		)

		BeforeEach(func() {
			buf.Reset()
			w = tarfs.NewWriter(&buf)
		})

		Describe("Read", func() {
			It("should return ErrPermission", func() {
				file, err := w.Create("test.txt")
				Expect(err).NotTo(HaveOccurred())
				DeferCleanup(file.Close)

				n, err := file.Read(make([]byte, 10))

				Expect(n).To(Equal(0))
				Expect(err).To(MatchError(ihfs.ErrPermission))
			})
		})

		Describe("Name", func() {
			It("should return the file name", func() {
				file, err := w.Create("test.txt")
				Expect(err).NotTo(HaveOccurred())
				DeferCleanup(file.Close)

				Expect(file).To(HaveField("Name()", "test.txt"))
			})
		})

		Describe("Stat", func() {
			It("should return ErrPermission", func() {
				file, err := w.Create("test.txt")
				Expect(err).NotTo(HaveOccurred())
				DeferCleanup(file.Close)

				info, err := file.Stat()

				Expect(info).To(BeNil())
				Expect(err).To(MatchError(ihfs.ErrPermission))
			})
		})

		Describe("Close", func() {
			It("should return error when writeEntry fails", func() {
				file, err := w.Create("test.txt")
				Expect(err).NotTo(HaveOccurred())

				_, err = file.(io.Writer).Write([]byte("content"))
				Expect(err).NotTo(HaveOccurred())

				// Close the Writer first so the subsequent writeEntry in file.Close fails
				Expect(w.Close()).To(Succeed())

				err = file.Close()
				Expect(err).To(HaveOccurred())
			})

			It("should be idempotent", func() {
				file, err := w.Create("test.txt")
				Expect(err).NotTo(HaveOccurred())

				Expect(file.Close()).To(Succeed())
				Expect(file.Close()).To(Succeed())

				Expect(w.Close()).To(Succeed())

				tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))
				entries, err := fs.ReadDir(tfs, ".")
				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(HaveLen(1))
			})
		})
	})

	Describe("Mkdir", func() {
		It("should write a directory entry", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			Expect(w.Mkdir("mydir", 0755)).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))
			f, err := tfs.Open("mydir")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(f.Close)
			info, err := f.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
		})

		It("should return ErrInvalid for an invalid path", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})

			err := w.Mkdir("../invalid", 0755)

			Expect(err).To(MatchError(ihfs.ErrInvalid))
		})

		It("should return error when WriteHeader fails", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})
			Expect(w.Close()).To(Succeed())

			err := w.Mkdir("mydir", 0755)

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("WriteFile", func() {
		It("should write file content as a tar entry", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			Expect(w.WriteFile("hello.txt", []byte("hello"), 0644)).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))
			f, err := tfs.Open("hello.txt")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(f.Close)
			content, err := io.ReadAll(f)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("hello"))
		})

		It("should return ErrInvalid for an invalid path", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})

			err := w.WriteFile("../invalid", []byte("data"), 0644)

			Expect(err).To(MatchError(ihfs.ErrInvalid))
		})

		It("should return error when WriteHeader fails", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})
			Expect(w.Close()).To(Succeed())

			err := w.WriteFile("test.txt", []byte("data"), 0644)

			Expect(err).To(HaveOccurred())
		})

		It("should return error when Write fails", func() {
			writeErr := errors.New("write failed")
			// Allow 3 writes (tar.WriteHeader makes 3 internal calls), then fail on the data write.
			fw := &failAfterWriter{w: &bytes.Buffer{}, n: 3, err: writeErr}
			w := tarfs.NewWriter(fw)

			err := w.WriteFile("test.txt", []byte("data"), 0644)

			Expect(err).To(MatchError(writeErr))
		})
	})
})

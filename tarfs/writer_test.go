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

// failAfterWriter allows n bytes to be written successfully, then returns err on all subsequent calls.
type failAfterWriter struct {
	w   io.Writer
	n   int
	err error
}

func (f *failAfterWriter) Write(p []byte) (int, error) {
	if f.n <= 0 {
		return 0, f.err
	}
	n, err := f.w.Write(p)
	f.n -= n
	return n, err
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

		Describe("Write", func() {
			It("should return ErrClosed after Close", func() {
				file, err := w.Create("test.txt")
				Expect(err).NotTo(HaveOccurred())
				Expect(file.Close()).To(Succeed())

				n, err := file.(io.Writer).Write([]byte("data"))

				Expect(n).To(Equal(0))
				Expect(err).To(MatchError(ihfs.ErrClosed))
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

			It("should return the same error on subsequent calls after a failed close", func() {
				file, err := w.Create("test.txt")
				Expect(err).NotTo(HaveOccurred())

				// Close the Writer first so the subsequent writeEntry in file.Close fails
				Expect(w.Close()).To(Succeed())

				firstErr := file.Close()
				Expect(firstErr).To(HaveOccurred())
				Expect(file.Close()).To(MatchError(firstErr))
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

	Describe("WriteEntry", func() {
		It("should write a header with no body when r is nil", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			hdr := &tar.Header{
				Typeflag: tar.TypeSymlink,
				Name:     "link.txt",
				Linkname: "target.txt",
			}
			Expect(w.WriteEntry(hdr, nil)).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tr := tar.NewReader(&buf)
			got, err := tr.Next()
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal("link.txt"))
			Expect(got.Linkname).To(Equal("target.txt"))
		})

		It("should write a header with body from r", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			content := []byte("hello")
			hdr := &tar.Header{
				Typeflag: tar.TypeReg,
				Name:     "file.txt",
				Size:     int64(len(content)),
			}
			Expect(w.WriteEntry(hdr, bytes.NewReader(content))).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))
			data, err := fs.ReadFile(tfs, "file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("hello"))
		})

		It("should return error when WriteHeader fails", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})
			Expect(w.Close()).To(Succeed())

			err := w.WriteEntry(&tar.Header{Name: "file.txt"}, nil)

			Expect(err).To(HaveOccurred())
		})

		It("should return error when io.Copy fails", func() {
			writeErr := errors.New("write failed")
			fw := &failAfterWriter{w: &bytes.Buffer{}, n: 512, err: writeErr}
			w := tarfs.NewWriter(fw)

			hdr := &tar.Header{
				Typeflag: tar.TypeReg,
				Name:     "file.txt",
				Size:     10,
			}
			err := w.WriteEntry(hdr, bytes.NewReader([]byte("0123456789")))

			Expect(err).To(MatchError(writeErr))
		})
	})

	Describe("Symlink", func() {
		It("should write a symlink entry", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			Expect(w.Symlink("link.txt", "target.txt")).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tr := tar.NewReader(&buf)
			hdr, err := tr.Next()
			Expect(err).NotTo(HaveOccurred())
			Expect(hdr.Typeflag).To(Equal(uint8(tar.TypeSymlink)))
			Expect(hdr.Name).To(Equal("link.txt"))
			Expect(hdr.Linkname).To(Equal("target.txt"))
		})

		It("should return ErrInvalid for an invalid path", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})

			err := w.Symlink("../invalid", "target.txt")

			Expect(err).To(MatchError(ihfs.ErrInvalid))
		})

		It("should return error when WriteHeader fails", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})
			Expect(w.Close()).To(Succeed())

			err := w.Symlink("link.txt", "target.txt")

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
			// Allow 512 bytes (one tar block written by WriteHeader), then fail on the data write.
			fw := &failAfterWriter{w: &bytes.Buffer{}, n: 512, err: writeErr}
			w := tarfs.NewWriter(fw)

			err := w.WriteFile("test.txt", []byte("data"), 0644)

			Expect(err).To(MatchError(writeErr))
		})
	})
})

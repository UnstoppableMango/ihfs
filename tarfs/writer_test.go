package tarfs_test

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/memfs"
	"github.com/unstoppablemango/ihfs/tarfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

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

	Describe("Copy", func() {
		It("should copy files with full metadata", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			m := memfs.New()
			Expect(m.Mkdir("subdir", 0755)).To(Succeed())
			f, err := m.Create("subdir/hello.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = f.(io.Writer).Write([]byte("hello"))
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			Expect(w.Copy(".", m)).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))
			data, err := fs.ReadFile(tfs, "subdir/hello.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("hello"))
		})

		It("should apply the dir prefix to all entries", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			m := memfs.New()
			Expect(m.WriteFile("app.bin", []byte("bin"), 0755)).To(Succeed())

			Expect(w.Copy("usr/local", m)).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tr := tar.NewReader(&buf)
			var names []string
			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break
				}
				Expect(err).NotTo(HaveOccurred())
				if hdr.Name != "." {
					names = append(names, hdr.Name)
				}
			}
			Expect(names).To(ContainElement("usr/local/app.bin"))
		})

		It("should copy symlinks", func() {
			entry := testfs.NewDirEntry("link.txt", false)
			entry.TypeFunc = func() ihfs.FileMode { return fs.ModeSymlink }
			entry.InfoFunc = func() (ihfs.FileInfo, error) {
				fi := testfs.NewFileInfo("link.txt")
				fi.ModeFunc = func() fs.FileMode { return fs.ModeSymlink }
				return fi, nil
			}
			fsys := testfs.New(
				testfs.WithStat(rootDirStat),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
				testfs.WithReadLink(func(string) (string, error) {
					return "target.txt", nil
				}),
			)

			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)
			Expect(w.Copy(".", fsys)).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tr := tar.NewReader(&buf)
			var found bool
			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break
				}
				Expect(err).NotTo(HaveOccurred())
				if hdr.Name == "link.txt" {
					found = true
					Expect(hdr.Typeflag).To(Equal(uint8(tar.TypeSymlink)))
					Expect(hdr.Linkname).To(Equal("target.txt"))
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should propagate walk errors", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})

			err := w.Copy(".", testfs.BoringFs{})

			Expect(err).To(HaveOccurred())
		})

		It("should propagate Info errors", func() {
			infoErr := errors.New("info error")
			entry := testfs.NewDirEntry("file.txt", false)
			entry.InfoFunc = func() (ihfs.FileInfo, error) { return nil, infoErr }
			fsys := testfs.New(
				testfs.WithStat(rootDirStat),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
			)

			err := tarfs.NewWriter(&bytes.Buffer{}).Copy(".", fsys)

			Expect(err).To(MatchError(infoErr))
		})

		It("should propagate ReadLink errors", func() {
			readLinkErr := errors.New("readlink error")
			entry := testfs.NewDirEntry("link.txt", false)
			entry.TypeFunc = func() ihfs.FileMode { return fs.ModeSymlink }
			entry.InfoFunc = func() (ihfs.FileInfo, error) {
				fi := testfs.NewFileInfo("link.txt")
				fi.ModeFunc = func() fs.FileMode { return fs.ModeSymlink }
				return fi, nil
			}
			fsys := testfs.New(
				testfs.WithStat(rootDirStat),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
				testfs.WithReadLink(func(string) (string, error) {
					return "", readLinkErr
				}),
			)

			err := tarfs.NewWriter(&bytes.Buffer{}).Copy(".", fsys)

			Expect(err).To(MatchError(readLinkErr))
		})

		It("should propagate FileInfoHeader errors for unsupported file types", func() {
			fi := testfs.NewFileInfo("socket.sock")
			fi.ModeFunc = func() ihfs.FileMode { return fs.ModeSocket }
			entry := testfs.NewDirEntry("socket.sock", false)
			entry.InfoFunc = func() (ihfs.FileInfo, error) { return fi, nil }
			fsys := testfs.New(
				testfs.WithStat(rootDirStat),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
			)

			err := tarfs.NewWriter(&bytes.Buffer{}).Copy(".", fsys)

			Expect(err).To(HaveOccurred())
		})

		It("should propagate Open errors for regular files", func() {
			openErr := errors.New("open error")
			entry := testfs.NewDirEntry("file.txt", false)
			fsys := testfs.New(
				testfs.WithStat(rootDirStat),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return nil, openErr
				}),
			)

			err := tarfs.NewWriter(&bytes.Buffer{}).Copy(".", fsys)

			Expect(err).To(MatchError(openErr))
		})

		It("should propagate WriteEntry errors", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})
			Expect(w.Close()).To(Succeed())

			err := w.Copy(".", memfs.New())

			Expect(err).To(HaveOccurred())
		})

		It("should return ErrInvalid for an invalid dir", func() {
			err := tarfs.NewWriter(&bytes.Buffer{}).Copy("../invalid", memfs.New())

			Expect(err).To(MatchError(ihfs.ErrInvalid))
		})

		It("should return ErrExist when copying the same file twice", func() {
			m := memfs.New()
			Expect(m.WriteFile("file.txt", []byte("data"), 0644)).To(Succeed())
			w := tarfs.NewWriter(&bytes.Buffer{})

			Expect(w.Copy(".", m)).To(Succeed())
			err := w.Copy(".", m)

			Expect(err).To(MatchError(fs.ErrExist))
		})
	})

	Describe("MkdirAll", func() {
		It("should write directory entries for each path component", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			Expect(w.MkdirAll("a/b/c", 0755)).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))
			for _, dir := range []string{"a", "a/b", "a/b/c"} {
				f, err := tfs.Open(dir)
				Expect(err).NotTo(HaveOccurred(), "expected %s to exist", dir)
				info, err := f.Stat()
				Expect(err).NotTo(HaveOccurred())
				Expect(info.IsDir()).To(BeTrue(), "expected %s to be a directory", dir)
				Expect(f.Close()).To(Succeed())
			}
		})

		It("should not write duplicate entries for already-created directories", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			Expect(w.MkdirAll("a/b", 0755)).To(Succeed())
			Expect(w.MkdirAll("a/b", 0755)).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tr := tar.NewReader(&buf)
			var names []string
			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break
				}
				Expect(err).NotTo(HaveOccurred())
				names = append(names, hdr.Name)
			}
			Expect(names).To(Equal([]string{"a/", "a/b/"}))
		})

		It("should do nothing for the root path", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			Expect(w.MkdirAll(".", 0755)).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tr := tar.NewReader(&buf)
			_, err := tr.Next()
			Expect(err).To(MatchError(io.EOF))
		})

		It("should return ErrInvalid for an invalid path", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})

			err := w.MkdirAll("../invalid", 0755)

			Expect(err).To(MatchError(ihfs.ErrInvalid))
		})

		It("should return error when WriteHeader fails", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})
			Expect(w.Close()).To(Succeed())

			err := w.MkdirAll("a/b", 0755)

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("OpenFile", func() {
		It("should return a writable file for O_WRONLY", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			file, err := w.OpenFile("test.txt", os.O_CREATE|os.O_WRONLY, 0644)

			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			DeferCleanup(file.Close)
		})

		It("should return a writable file for O_RDWR", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			file, err := w.OpenFile("test.txt", os.O_CREATE|os.O_RDWR, 0644)

			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			DeferCleanup(file.Close)
		})

		It("should write content as a tar entry on Close", func() {
			var buf bytes.Buffer
			w := tarfs.NewWriter(&buf)

			file, err := w.OpenFile("hello.txt", os.O_CREATE|os.O_WRONLY, 0644)
			Expect(err).NotTo(HaveOccurred())
			_, err = file.(io.Writer).Write([]byte("hello"))
			Expect(err).NotTo(HaveOccurred())
			Expect(file.Close()).To(Succeed())
			Expect(w.Close()).To(Succeed())

			tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))
			data, err := fs.ReadFile(tfs, "hello.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("hello"))
		})

		It("should return ErrPermission for O_RDONLY", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})

			file, err := w.OpenFile("test.txt", os.O_RDONLY, 0644)

			Expect(file).To(BeNil())
			Expect(err).To(MatchError(ihfs.ErrPermission))
		})

		It("should return ErrInvalid for an invalid path", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})

			file, err := w.OpenFile("../invalid", os.O_CREATE|os.O_WRONLY, 0644)

			Expect(file).To(BeNil())
			Expect(err).To(MatchError(ihfs.ErrInvalid))
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

			Expect(w.Symlink("target.txt", "link.txt")).To(Succeed())
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

			err := w.Symlink("target.txt", "../invalid")

			Expect(err).To(MatchError(ihfs.ErrInvalid))
		})

		It("should return error when WriteHeader fails", func() {
			w := tarfs.NewWriter(&bytes.Buffer{})
			Expect(w.Close()).To(Succeed())

			err := w.Symlink("target.txt", "link.txt")

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

package prefixfs_test

import (
	"io"
	"io/fs"
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/prefixfs"
)

var _ = Describe("Fs", func() {
	var inner fstest.MapFS

	BeforeEach(func() {
		inner = fstest.MapFS{
			"file.txt":        &fstest.MapFile{Data: []byte("hello")},
			"subdir/deep.txt": &fstest.MapFile{Data: []byte("world")},
		}
	})

	Describe("New", func() {
		It("should create an Fs with a single-level prefix", func() {
			f, err := prefixfs.New(inner, "a")
			Expect(err).NotTo(HaveOccurred())
			Expect(f).NotTo(BeNil())
		})

		It("should create an Fs with a multi-level prefix", func() {
			f, err := prefixfs.New(inner, "a/b/c")
			Expect(err).NotTo(HaveOccurred())
			Expect(f).NotTo(BeNil())
		})

		It("should clean the prefix before validating", func() {
			f, err := prefixfs.New(inner, "a/b/../c")
			Expect(err).NotTo(HaveOccurred())
			Expect(f).NotTo(BeNil())
		})

		It("should return ErrInvalid for '.'", func() {
			_, err := prefixfs.New(inner, ".")
			Expect(err).To(MatchError(fs.ErrInvalid))
		})

		It("should return ErrInvalid for an absolute path", func() {
			_, err := prefixfs.New(inner, "/absolute")
			Expect(err).To(MatchError(fs.ErrInvalid))
		})
	})

	Describe("Base", func() {
		It("should return the wrapped FS", func() {
			f, err := prefixfs.New(inner, "a")
			Expect(err).NotTo(HaveOccurred())

			Expect(f.Base()).To(Equal(ihfs.FS(inner)))
		})
	})

	Describe("Open", func() {
		var fsys *prefixfs.Fs

		BeforeEach(func() {
			var err error
			fsys, err = prefixfs.New(inner, "a/b")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return ErrInvalid for invalid paths", func() {
			_, err := fsys.Open("")
			Expect(err).To(MatchError(fs.ErrInvalid))
		})

		It("should open the virtual root", func() {
			f, err := fsys.Open(".")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(f.Close)
		})

		It("should open a virtual ancestor directory", func() {
			f, err := fsys.Open("a")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(f.Close)
		})

		It("should open the prefix path as the underlying root", func() {
			f, err := fsys.Open("a/b")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(f.Close)
		})

		It("should open a file under the prefix", func() {
			f, err := fsys.Open("a/b/file.txt")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(f.Close)
		})

		It("should return ErrNotExist for sibling paths", func() {
			_, err := fsys.Open("other")
			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("should implement ihfs.FS", func() {
			var _ ihfs.FS = fsys
		})
	})

	Describe("Stat", func() {
		var fsys *prefixfs.Fs

		BeforeEach(func() {
			var err error
			fsys, err = prefixfs.New(inner, "a/b")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return ErrInvalid for invalid paths", func() {
			_, err := fsys.Stat("")
			Expect(err).To(MatchError(fs.ErrInvalid))
		})

		It("should stat the virtual root", func() {
			info, err := fsys.Stat(".")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
			Expect(info.Name()).To(Equal("."))
			Expect(info.Size()).To(BeZero())
			Expect(info.Mode()).To(Equal(fs.ModeDir | 0o555))
			Expect(info.ModTime()).To(BeZero())
			Expect(info.Sys()).To(BeNil())
		})

		It("should stat a virtual ancestor directory", func() {
			info, err := fsys.Stat("a")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
			Expect(info.Name()).To(Equal("a"))
		})

		It("should stat the prefix path as the underlying root", func() {
			info, err := fsys.Stat("a/b")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
		})

		It("should stat a file under the prefix", func() {
			info, err := fsys.Stat("a/b/file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeFalse())
			Expect(info.Name()).To(Equal("file.txt"))
		})

		It("should return ErrNotExist for sibling paths", func() {
			_, err := fsys.Stat("other")
			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("should implement ihfs.StatFS", func() {
			var _ ihfs.StatFS = fsys
		})
	})

	Describe("ReadDir", func() {
		var fsys *prefixfs.Fs

		BeforeEach(func() {
			var err error
			fsys, err = prefixfs.New(inner, "a/b")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return ErrInvalid for invalid paths", func() {
			_, err := fsys.ReadDir("")
			Expect(err).To(MatchError(fs.ErrInvalid))
		})

		It("should read the virtual root and return its single child", func() {
			entries, err := fsys.ReadDir(".")
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("a"))
			Expect(entries[0].IsDir()).To(BeTrue())
		})

		It("should read a virtual ancestor directory and return its single child", func() {
			entries, err := fsys.ReadDir("a")
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("b"))
			Expect(entries[0].IsDir()).To(BeTrue())
		})

		It("should read the prefix path as the underlying root", func() {
			entries, err := fsys.ReadDir("a/b")
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).NotTo(BeEmpty())
		})

		It("should read a directory under the prefix", func() {
			entries, err := fsys.ReadDir("a/b/subdir")
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("deep.txt"))
		})

		It("should return ErrNotExist for sibling paths", func() {
			_, err := fsys.ReadDir("other")
			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("should implement ihfs.ReadDirFS", func() {
			var _ ihfs.ReadDirFS = fsys
		})
	})

	Describe("virtual root directory file", func() {
		var (
			fsys *prefixfs.Fs
			dir  fs.File
		)

		BeforeEach(func() {
			var err error
			fsys, err = prefixfs.New(inner, "a/b")
			Expect(err).NotTo(HaveOccurred())
			dir, err = fsys.Open(".")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(dir.Close)
		})

		It("should stat as a directory with name '.'", func() {
			info, err := dir.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
			Expect(info.Name()).To(Equal("."))
		})

		It("should fail to read bytes", func() {
			_, err := dir.Read(make([]byte, 10))
			Expect(err).To(HaveOccurred())
		})

		It("should report the full path in read errors for multi-level ancestors", func() {
			deep, err := prefixfs.New(inner, "a/b/c")
			Expect(err).NotTo(HaveOccurred())

			f, err := deep.Open("a/b")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(f.Close)

			_, err = f.Read(make([]byte, 10))
			var pathErr *fs.PathError
			Expect(err).To(BeAssignableToTypeOf(pathErr))
			Expect(err.(*fs.PathError).Path).To(Equal("a/b"))
		})

		It("should implement fs.ReadDirFile", func() {
			_, ok := dir.(fs.ReadDirFile)
			Expect(ok).To(BeTrue())
		})

		It("should read all entries when n <= 0", func() {
			rdf := dir.(fs.ReadDirFile)
			entries, err := rdf.ReadDir(-1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("a"))
		})

		It("should return nil when n <= 0 and entries are exhausted", func() {
			rdf := dir.(fs.ReadDirFile)
			_, _ = rdf.ReadDir(-1)
			entries, err := rdf.ReadDir(-1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(BeEmpty())
		})

		It("should read entries when n > 0", func() {
			rdf := dir.(fs.ReadDirFile)
			entries, err := rdf.ReadDir(1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
		})

		It("should return io.EOF when n > 0 and entries are exhausted", func() {
			rdf := dir.(fs.ReadDirFile)
			_, _ = rdf.ReadDir(1)
			_, err := rdf.ReadDir(1)
			Expect(err).To(MatchError(io.EOF))
		})
	})

	Describe("virtual ancestor directory file", func() {
		var (
			fsys *prefixfs.Fs
			dir  fs.File
		)

		BeforeEach(func() {
			var err error
			fsys, err = prefixfs.New(inner, "a/b/c")
			Expect(err).NotTo(HaveOccurred())
			dir, err = fsys.Open("a")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(dir.Close)
		})

		It("should stat as a directory with the correct name", func() {
			info, err := dir.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
			Expect(info.Name()).To(Equal("a"))
		})

		It("should read its single child", func() {
			rdf := dir.(fs.ReadDirFile)
			entries, err := rdf.ReadDir(-1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("b"))
		})
	})

	Describe("virtual DirEntry", func() {
		var (
			fsys    *prefixfs.Fs
			entries []fs.DirEntry
		)

		BeforeEach(func() {
			var err error
			fsys, err = prefixfs.New(inner, "a/b")
			Expect(err).NotTo(HaveOccurred())
			entries, err = fsys.ReadDir(".")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should have the correct name", func() {
			Expect(entries[0].Name()).To(Equal("a"))
		})

		It("should report IsDir true", func() {
			Expect(entries[0].IsDir()).To(BeTrue())
		})

		It("should report Type as ModeDir", func() {
			Expect(entries[0].Type()).To(Equal(fs.ModeDir))
		})

		It("should return FileInfo via Info", func() {
			info, err := entries[0].Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Name()).To(Equal("a"))
			Expect(info.IsDir()).To(BeTrue())
		})

		It("should not implement fs.ReadDirFile", func() {
			_, ok := entries[0].(fs.ReadDirFile)
			Expect(ok).To(BeFalse())
		})
	})

	Describe("single-level prefix", func() {
		var fsys *prefixfs.Fs

		BeforeEach(func() {
			var err error
			fsys, err = prefixfs.New(inner, "mount")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should open files under the prefix", func() {
			f, err := fsys.Open("mount/file.txt")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(f.Close)
		})

		It("should read the root with the mount point as child", func() {
			entries, err := fsys.ReadDir(".")
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("mount"))
		})
	})
})

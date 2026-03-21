package layer_test

import (
	"archive/tar"
	"errors"
	"io"
	"io/fs"
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/ctrfs/v1/layer"
	"github.com/unstoppablemango/ihfs/memfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("FS", func() {
	Describe("From", func() {
		It("should open a file from a layer", func() {
			l, err := makeLayer([]tarEntry{
				{hdr: &tar.Header{Name: "hello.txt", Typeflag: tar.TypeReg, Size: 11, Mode: 0644}, data: "hello world"},
			})
			Expect(err).NotTo(HaveOccurred())

			fsys, err := layer.From(l)
			Expect(err).NotTo(HaveOccurred())
			defer fsys.Close()

			data, err := fs.ReadFile(fsys, "hello.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("hello world"))
		})

		It("should stat a file from a layer", func() {
			l, err := makeLayer([]tarEntry{
				{hdr: &tar.Header{Name: "info.txt", Typeflag: tar.TypeReg, Size: 4, Mode: 0644}, data: "info"},
			})
			Expect(err).NotTo(HaveOccurred())

			fsys, err := layer.From(l)
			Expect(err).NotTo(HaveOccurred())
			defer fsys.Close()

			info, err := fs.Stat(fsys, "info.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Name()).To(Equal("info.txt"))
			Expect(info.IsDir()).To(BeFalse())
		})

		Describe("fstest", func() {
			It("should pass fstest.TestFS", func() {
				l, err := makeLayer([]tarEntry{
					{hdr: &tar.Header{Name: "dir/", Typeflag: tar.TypeDir, Mode: 0755}},
					{hdr: &tar.Header{Name: "dir/hello.txt", Typeflag: tar.TypeReg, Size: 5, Mode: 0644}, data: "hello"},
					{hdr: &tar.Header{Name: "readme.md", Typeflag: tar.TypeReg, Size: 6, Mode: 0644}, data: "readme"},
				})
				Expect(err).NotTo(HaveOccurred())

				fsys, err := layer.From(l)
				Expect(err).NotTo(HaveOccurred())
				defer fsys.Close()

				err = fstest.TestFS(fsys, "dir/hello.txt", "readme.md")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("should return error when Uncompressed fails", func() {
			fsys, err := layer.From(&errLayer{err: errors.New("uncompressed error")})

			Expect(err).To(MatchError("uncompressed error"))
			Expect(fsys).To(BeNil())
		})
	})
})

var _ = Describe("Create", func() {
	It("should create a layer from an fs.FS", func() {
		m := memfs.New()
		Expect(m.Mkdir("subdir", 0755)).To(Succeed())
		f, err := m.Create("subdir/hello.txt")
		Expect(err).NotTo(HaveOccurred())
		_, err = f.(io.Writer).Write([]byte("hello"))
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Close()).To(Succeed())

		l, err := layer.Create(m, ".")
		Expect(err).NotTo(HaveOccurred())

		rc, err := l.Uncompressed()
		Expect(err).NotTo(HaveOccurred())
		defer rc.Close()

		names, err := tarNames(rc)
		Expect(err).NotTo(HaveOccurred())
		Expect(names).To(ContainElements("subdir/", "subdir/hello.txt"))
	})

	It("should create a layer rooted at a subdirectory", func() {
		m := memfs.New()
		Expect(m.Mkdir("src", 0755)).To(Succeed())
		f, err := m.Create("src/main.go")
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Close()).To(Succeed())

		l, err := layer.Create(m, "src")
		Expect(err).NotTo(HaveOccurred())

		rc, err := l.Uncompressed()
		Expect(err).NotTo(HaveOccurred())
		defer rc.Close()

		names, err := tarNames(rc)
		Expect(err).NotTo(HaveOccurred())
		Expect(names).To(ContainElement("main.go"))
	})

	It("should include symlinks in the layer", func() {
		entry := testfs.NewDirEntry("link.txt", false)
		entry.TypeFunc = func() ihfs.FileMode { return ihfs.FileMode(fs.ModeSymlink) }
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

		l, err := layer.Create(fsys, ".")
		Expect(err).NotTo(HaveOccurred())

		rc, err := l.Uncompressed()
		Expect(err).NotTo(HaveOccurred())
		defer rc.Close()

		names, err := tarNames(rc)
		Expect(err).NotTo(HaveOccurred())
		Expect(names).To(ContainElement("link.txt"))
	})

	It("should propagate ReadLink errors for symlinks", func() {
		readLinkErr := errors.New("readlink error")
		entry := testfs.NewDirEntry("link.txt", false)
		entry.TypeFunc = func() ihfs.FileMode { return ihfs.FileMode(fs.ModeSymlink) }
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

		_, err := layer.Create(fsys, ".")

		Expect(err).To(MatchError(readLinkErr))
	})

	It("should propagate walk errors", func() {
		_, err := layer.Create(testfs.BoringFs{}, "nonexistent")

		Expect(err).To(HaveOccurred())
	})

	It("should propagate FileInfoHeader errors for unsupported file modes", func() {
		fi := testfs.NewFileInfo("socket.sock")
		fi.ModeFunc = func() ihfs.FileMode { return ihfs.FileMode(fs.ModeSocket) }
		entry := testfs.NewDirEntry("socket.sock", false)
		entry.InfoFunc = func() (ihfs.FileInfo, error) { return fi, nil }
		fsys := testfs.New(
			testfs.WithStat(rootDirStat),
			testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
				return []ihfs.DirEntry{entry}, nil
			}),
		)

		_, err := layer.Create(fsys, ".")

		Expect(err).To(HaveOccurred())
	})

	It("should propagate Info errors from WalkDir", func() {
		infoErr := errors.New("info error")
		entry := testfs.NewDirEntry("file.txt", false)
		entry.InfoFunc = func() (ihfs.FileInfo, error) { return nil, infoErr }
		fsys := testfs.New(
			testfs.WithStat(rootDirStat),
			testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
				return []ihfs.DirEntry{entry}, nil
			}),
		)

		_, err := layer.Create(fsys, ".")

		Expect(err).To(MatchError(infoErr))
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

		_, err := layer.Create(fsys, ".")

		Expect(err).To(MatchError(openErr))
	})

	It("should propagate io.Copy errors", func() {
		copyErr := errors.New("read error")
		srcFile := &testfs.File{
			StatFunc:  func() (ihfs.FileInfo, error) { return testfs.NewFileInfo("file.txt"), nil },
			CloseFunc: func() error { return nil },
			ReadFunc:  func([]byte) (int, error) { return 0, copyErr },
		}
		entry := testfs.NewDirEntry("file.txt", false)
		fsys := testfs.New(
			testfs.WithStat(rootDirStat),
			testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
				return []ihfs.DirEntry{entry}, nil
			}),
			testfs.WithOpen(func(string) (ihfs.File, error) {
				return srcFile, nil
			}),
		)

		_, err := layer.Create(fsys, ".")

		Expect(err).To(MatchError(copyErr))
	})
})

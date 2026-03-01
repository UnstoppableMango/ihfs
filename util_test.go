package ihfs_test

import (
	"bytes"
	"errors"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/osfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("Util", func() {
	Describe("DirExists", func() {
		It("should return true when path is a directory", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				fi := testfs.NewFileInfo(s)
				fi.IsDirFunc = func() bool { return true }
				return fi, nil
			}))

			exists, err := ihfs.DirExists(fsys, "dir")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("should return false when path is a file", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return testfs.NewFileInfo(s), nil
			}))

			exists, err := ihfs.DirExists(fsys, "file.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("should return false when path does not exist", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return nil, fs.ErrNotExist
			}))

			exists, err := ihfs.DirExists(fsys, "nonexistent")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("should return error when stat returns an error", func() {
			testErr := errors.New("test error")
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return nil, testErr
			}))

			exists, err := ihfs.DirExists(fsys, "dir")

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(testErr))
			Expect(exists).To(BeFalse())
		})
	})

	Describe("Exists", func() {
		It("should return true for file", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return testfs.NewFileInfo(s), nil
			}))

			exists, err := ihfs.Exists(fsys, "file.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("should return true for directory", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				fi := testfs.NewFileInfo(s)
				fi.IsDirFunc = func() bool { return true }
				return fi, nil
			}))

			exists, err := ihfs.Exists(fsys, "dir")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("should return false when path does not exist", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return nil, fs.ErrNotExist
			}))

			exists, err := ihfs.Exists(fsys, "nonexistent")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("should return error when stat returns an error", func() {
			testErr := errors.New("test error")
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return nil, testErr
			}))

			exists, err := ihfs.Exists(fsys, "file.txt")

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(testErr))
			Expect(exists).To(BeFalse())
		})
	})

	Describe("IsDir", func() {
		It("should return true when path is a directory", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				fi := testfs.NewFileInfo(s)
				fi.IsDirFunc = func() bool { return true }
				return fi, nil
			}))

			isDir, err := ihfs.IsDir(fsys, "dir")

			Expect(err).NotTo(HaveOccurred())
			Expect(isDir).To(BeTrue())
		})

		It("should return false when path is a file", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return testfs.NewFileInfo(s), nil
			}))

			isDir, err := ihfs.IsDir(fsys, "file.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(isDir).To(BeFalse())
		})

		It("should return error when path does not exist", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return nil, fs.ErrNotExist
			}))

			isDir, err := ihfs.IsDir(fsys, "nonexistent")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fs.ErrNotExist))
			Expect(isDir).To(BeFalse())
		})

		It("should return error when stat returns an error", func() {
			testErr := errors.New("test error")
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return nil, testErr
			}))

			isDir, err := ihfs.IsDir(fsys, "dir")

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(testErr))
			Expect(isDir).To(BeFalse())
		})
	})

	Describe("ReadDirNames", func() {
		It("should read directory entry names", func() {
			fsys := osfs.New()

			names, err := ihfs.ReadDirNames(fsys, "./testdata/2-files")

			Expect(err).NotTo(HaveOccurred())
			Expect(names).To(ConsistOf("one.txt", "two.txt"))
		})

		It("should return error when directory does not exist", func() {
			fsys := osfs.New()

			names, err := ihfs.ReadDirNames(fsys, "./nonexistent")

			Expect(err).To(HaveOccurred())
			Expect(names).To(BeNil())
		})
	})

	Describe("WriteReader", func() {
		It("should write reader contents to file", func() {
			var capturedName string
			var capturedData []byte
			var capturedPerm ihfs.FileMode

			fsys := testfs.New(testfs.WithWriteFile(func(name string, data []byte, perm ihfs.FileMode) error {
				capturedName = name
				capturedData = data
				capturedPerm = perm
				return nil
			}))

			reader := bytes.NewReader([]byte("test content"))
			err := ihfs.WriteReader(fsys, "test.txt", reader, 0x644)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("test.txt"))
			Expect(capturedData).To(Equal([]byte("test content")))
			Expect(capturedPerm).To(Equal(ihfs.FileMode(0x644)))
		})

		It("should return error when reading fails", func() {
			fsys := testfs.New(testfs.WithWriteFile(func(name string, data []byte, perm ihfs.FileMode) error {
				return nil
			}))

			reader := &errorReader{err: errors.New("read error")}
			err := ihfs.WriteReader(fsys, "test.txt", reader, 0x644)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("reading"))
			Expect(err.Error()).To(ContainSubstring("read error"))
		})

		It("should return error when WriteFile fails", func() {
			writeErr := errors.New("write error")
			fsys := testfs.New(testfs.WithWriteFile(func(name string, data []byte, perm ihfs.FileMode) error {
				return writeErr
			}))

			reader := bytes.NewReader([]byte("test content"))
			err := ihfs.WriteReader(fsys, "test.txt", reader, 0x644)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(writeErr))
		})
	})

	Describe("Mkdir", func() {
		It("should call underlying Mkdir when MkdirFS is implemented", func() {
			var capturedPath string
			var capturedPerm ihfs.FileMode

			fsys := testfs.New(testfs.WithMkdir(func(path string, perm ihfs.FileMode) error {
				capturedPath = path
				capturedPerm = perm
				return nil
			}))

			err := ihfs.Mkdir(fsys, "testdir", 0o755)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedPath).To(Equal("testdir"))
			Expect(capturedPerm).To(Equal(ihfs.FileMode(0o755)))
		})

		It("should return ErrNotImplemented when MkdirFS not implemented", func() {
			fsys := testfs.BoringFs{}

			err := ihfs.Mkdir(fsys, "testdir", 0o755)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
		})

		It("should propagate errors from underlying Mkdir", func() {
			mkdirErr := errors.New("mkdir error")
			fsys := testfs.New(testfs.WithMkdir(func(path string, perm ihfs.FileMode) error {
				return mkdirErr
			}))

			err := ihfs.Mkdir(fsys, "testdir", 0o755)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(mkdirErr))
		})
	})

	Describe("MkdirAll", func() {
		It("should call underlying MkdirAll when MkdirAllFS is implemented", func() {
			var capturedPath string
			var capturedPerm ihfs.FileMode

			fsys := testfs.New(testfs.WithMkdirAll(func(path string, perm ihfs.FileMode) error {
				capturedPath = path
				capturedPerm = perm
				return nil
			}))

			err := ihfs.MkdirAll(fsys, "parent/child", 0o755)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedPath).To(Equal("parent/child"))
			Expect(capturedPerm).To(Equal(ihfs.FileMode(0o755)))
		})

		It("should return nil for empty path", func() {
			fsys := testfs.BoringFs{}

			err := ihfs.MkdirAll(fsys, "", 0o755)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should create directory when Mkdir succeeds", func() {
			var capturedPath string

			fsys := &mkdirOnlyFS{
				mkdirFunc: func(path string, perm ihfs.FileMode) error {
					capturedPath = path
					return nil
				},
			}

			err := ihfs.MkdirAll(fsys, "testdir", 0o755)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedPath).To(Equal("testdir"))
		})

		It("should create parent directories recursively", func() {
			var createdDirs []string
			callCount := 0

			fsys := &mkdirOnlyFS{
				mkdirFunc: func(path string, perm ihfs.FileMode) error {
					callCount++
					if callCount == 1 && path == "parent/child" {
						return fs.ErrNotExist
					}
					createdDirs = append(createdDirs, path)
					return nil
				},
			}

			err := ihfs.MkdirAll(fsys, "parent/child", 0o755)

			Expect(err).NotTo(HaveOccurred())
			Expect(createdDirs).To(ContainElement("parent"))
			Expect(createdDirs).To(ContainElement("parent/child"))
		})

		It("should stop at root when parent equals current path", func() {
			fsys := &mkdirOnlyFS{
				mkdirFunc: func(path string, perm ihfs.FileMode) error {
					return fs.ErrNotExist
				},
			}

			err := ihfs.MkdirAll(fsys, "/", 0o755)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, fs.ErrNotExist)).To(BeTrue())
		})

		It("should return error when Mkdir fails with non-ErrNotExist", func() {
			mkdirErr := errors.New("permission denied")
			fsys := &mkdirOnlyFS{
				mkdirFunc: func(path string, perm ihfs.FileMode) error {
					return mkdirErr
				},
			}

			err := ihfs.MkdirAll(fsys, "testdir", 0o755)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(mkdirErr))
		})

		It("should return error when creating parent fails", func() {
			parentErr := errors.New("parent mkdir failed")
			callCount := 0

			fsys := &mkdirOnlyFS{
				mkdirFunc: func(path string, perm ihfs.FileMode) error {
					callCount++
					if callCount == 1 && path == "parent/child" {
						return fs.ErrNotExist
					}
					if path == "parent" {
						return parentErr
					}
					return nil
				},
			}

			err := ihfs.MkdirAll(fsys, "parent/child", 0o755)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(parentErr))
		})

		It("should propagate errors from underlying MkdirAll", func() {
			mkdirAllErr := errors.New("mkdirall error")
			fsys := testfs.New(testfs.WithMkdirAll(func(path string, perm ihfs.FileMode) error {
				return mkdirAllErr
			}))

			err := ihfs.MkdirAll(fsys, "parent/child", 0o755)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(mkdirAllErr))
		})
	})

	Describe("WriteFile", func() {
		It("should call underlying WriteFile when implemented", func() {
			var capturedName string
			var capturedData []byte
			var capturedPerm ihfs.FileMode

			fsys := testfs.New(testfs.WithWriteFile(func(name string, data []byte, perm ihfs.FileMode) error {
				capturedName = name
				capturedData = data
				capturedPerm = perm
				return nil
			}))

			err := ihfs.WriteFile(fsys, "test.txt", []byte("content"), 0x644)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("test.txt"))
			Expect(capturedData).To(Equal([]byte("content")))
			Expect(capturedPerm).To(Equal(ihfs.FileMode(0x644)))
		})

		It("should return ErrNotImplemented when WriteFileFS not implemented", func() {
			fsys := testfs.BoringFs{}

			err := ihfs.WriteFile(fsys, "test.txt", []byte("content"), 0x644)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
		})
	})
})

type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

type mkdirOnlyFS struct {
	mkdirFunc func(string, ihfs.FileMode) error
}

func (m *mkdirOnlyFS) Open(name string) (ihfs.File, error) {
	return nil, fs.ErrNotExist
}

func (m *mkdirOnlyFS) Mkdir(name string, mode ihfs.FileMode) error {
	return m.mkdirFunc(name, mode)
}

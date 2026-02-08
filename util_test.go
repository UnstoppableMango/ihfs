package ihfs_test

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/osfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("Util", func() {
	Describe("Copy", func() {
		var tempDir string

		BeforeEach(func() {
			tempDir = GinkgoT().TempDir()
		})

		It("should copy files from source filesystem to directory", func() {
			srcFs, err := fs.Sub(osfs.New(), "testdata/2-files")
			Expect(err).NotTo(HaveOccurred())
			destDir := filepath.Join(tempDir, "dest")

			err = ihfs.Copy(destDir, srcFs)

			Expect(err).NotTo(HaveOccurred())

			// Verify files were copied
			oneContent, err := os.ReadFile(filepath.Join(destDir, "one.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(oneContent)).To(Equal("one\n"))

			twoContent, err := os.ReadFile(filepath.Join(destDir, "two.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(twoContent)).To(Equal("two\n"))
		})

		It("should create destination directory if it does not exist", func() {
			// Create a temp source directory with a file
			srcDir := filepath.Join(tempDir, "src")
			os.MkdirAll(srcDir, 0755)
			os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("content"), 0644)

			// Change to temp dir so we can use relative paths with Sub
			oldWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldWd)

			srcFs, err := fs.Sub(osfs.New(), "src")
			Expect(err).NotTo(HaveOccurred())
			destDir := filepath.Join(tempDir, "new-dir")

			err = ihfs.Copy(destDir, srcFs)

			Expect(err).NotTo(HaveOccurred())
			_, err = os.Stat(destDir)
			Expect(err).NotTo(HaveOccurred())
			content, err := os.ReadFile(filepath.Join(destDir, "test.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("content"))
		})

		It("should return error if file already exists", func() {
			// Create a temp source directory with a file
			srcDir := filepath.Join(tempDir, "src2")
			os.MkdirAll(srcDir, 0755)
			os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("content"), 0644)

			// Change to temp dir so we can use relative paths with Sub
			oldWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldWd)

			srcFs, err := fs.Sub(osfs.New(), "src2")
			Expect(err).NotTo(HaveOccurred())

			// Create destination file first
			destDir := filepath.Join(tempDir, "dest")
			os.MkdirAll(destDir, 0755)
			os.WriteFile(filepath.Join(destDir, "test.txt"), []byte("existing"), 0644)

			err = ihfs.Copy(destDir, srcFs)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, fs.ErrExist)).To(BeTrue())
		})

		It("should handle directory creation", func() {
			// Create a temp source directory with subdirectory
			srcDir := filepath.Join(tempDir, "src3")
			os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)
			os.WriteFile(filepath.Join(srcDir, "subdir", "file.txt"), []byte("content"), 0644)

			// Change to temp dir so we can use relative paths with Sub
			oldWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldWd)

			srcFs, err := fs.Sub(osfs.New(), "src3")
			Expect(err).NotTo(HaveOccurred())
			destDir := filepath.Join(tempDir, "dest")

			err = ihfs.Copy(destDir, srcFs)

			Expect(err).NotTo(HaveOccurred())
			info, err := os.Stat(filepath.Join(destDir, "subdir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
		})

		It("should use CopyFS interface when available", func() {
			var capturedDir string
			var called bool
			copyFs := testfs.New(
				testfs.WithCopy(func(dir string, src ihfs.FS) error {
					capturedDir = dir
					called = true
					return nil
				}),
			)

			err := ihfs.Copy("target", copyFs)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedDir).To(Equal("target"))
			Expect(called).To(BeTrue())
		})

		It("should return error from CopyFS interface", func() {
			copyErr := errors.New("copy failed")
			copyFs := testfs.New(
				testfs.WithCopy(func(dir string, src ihfs.FS) error {
					return copyErr
				}),
			)

			err := ihfs.Copy("target", copyFs)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(copyErr))
		})

		It("should return error when WalkDir encounters error during traversal", func() {
			// Create a temp source directory with a file
			srcDir := filepath.Join(tempDir, "src4")
			os.MkdirAll(srcDir, 0755)
			os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("content"), 0644)

			// Change to temp dir so we can use relative paths with Sub
			oldWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldWd)

			srcFs, err := fs.Sub(osfs.New(), "src4")
			Expect(err).NotTo(HaveOccurred())
			// Create an invalid destDir that will cause os.MkdirAll to fail
			// Use a file path as directory to trigger error
			destFile := filepath.Join(tempDir, "file-not-dir")
			os.WriteFile(destFile, []byte("blocking"), 0644)
			destDir := filepath.Join(destFile, "subdir")

			err = ihfs.Copy(destDir, srcFs)

			Expect(err).To(HaveOccurred())
		})

		It("should return error when Open returns error", func() {
			// Create a custom FS that fails on Open
			srcFs := &failingOpenFS{}

			err := ihfs.Copy(tempDir, srcFs)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("open failed"))
		})

		It("should clean up on copy error", func() {
			// Create a custom FS that fails during Read
			srcFs := &failingReadFS{}
			destDir := filepath.Join(tempDir, "dest")
			os.MkdirAll(destDir, 0755)

			err := ihfs.Copy(destDir, srcFs)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("read failed"))
			// Verify file was cleaned up
			_, err = os.Stat(filepath.Join(destDir, "test.txt"))
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("should propagate walkdir function error", func() {
			// Create a temp source directory with a file
			srcDir := filepath.Join(tempDir, "src5")
			os.MkdirAll(srcDir, 0755)
			os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("content"), 0644)

			// Change to temp dir so we can use relative paths with Sub
			oldWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldWd)

			srcFs, err := fs.Sub(osfs.New(), "src5")
			Expect(err).NotTo(HaveOccurred())
			// Use invalid permissions that will cause file creation error
			destDir := filepath.Join(tempDir, "dest2")
			os.MkdirAll(destDir, 0755)
			os.WriteFile(filepath.Join(destDir, "test.txt"), []byte("exists"), 0644)

			err = ihfs.Copy(destDir, srcFs)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, fs.ErrExist)).To(BeTrue())
		})

		It("should return error when DirEntry.Info() fails for directory", func() {
			srcFs := &failingInfoDirFS{}

			err := ihfs.Copy(tempDir, srcFs)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("info failed"))
		})

		It("should return error when DirEntry.Info() fails for file", func() {
			srcFs := &failingInfoFileFS{}

			err := ihfs.Copy(tempDir, srcFs)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("info failed"))
		})

		It("should return error when Walk passes error to callback", func() {
			// Create a directory that will cause walk to fail
			srcDir := filepath.Join(tempDir, "src6")
			subdir := filepath.Join(srcDir, "badperms")
			os.MkdirAll(subdir, 0755)
			os.WriteFile(filepath.Join(subdir, "file.txt"), []byte("content"), 0644)
			// Remove read permissions to cause walk error
			os.Chmod(subdir, 0000)
			defer os.Chmod(subdir, 0755) // Clean up

			// Change to temp dir so we can use relative paths with Sub
			oldWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldWd)

			srcFs, err := fs.Sub(osfs.New(), "src6")
			Expect(err).NotTo(HaveOccurred())
			destDir := filepath.Join(tempDir, "dest3")

			err = ihfs.Copy(destDir, srcFs)

			Expect(err).To(HaveOccurred())
		})

		It("should return error when OpenFile fails", func() {
			// Create a source file
			srcDir := filepath.Join(tempDir, "src7")
			os.MkdirAll(srcDir, 0755)
			os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("content"), 0644)

			// Change to temp dir
			oldWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldWd)

			srcFs, err := fs.Sub(osfs.New(), "src7")
			Expect(err).NotTo(HaveOccurred())

			// Create dest directory with no write permissions
			destDir := filepath.Join(tempDir, "dest4")
			os.MkdirAll(destDir, 0555)
			defer os.Chmod(destDir, 0755) // Clean up

			err = ihfs.Copy(destDir, srcFs)

			Expect(err).To(HaveOccurred())
		})
	})

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
			fsys := testfs.New(testfs.WithStat(func(string) (ihfs.FileInfo, error) {
				return nil, fs.ErrNotExist
			}))

			exists, err := ihfs.DirExists(fsys, "nonexistent")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("should return error when stat returns an error", func() {
			testErr := errors.New("test error")
			fsys := testfs.New(testfs.WithStat(func(string) (ihfs.FileInfo, error) {
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
			fsys := testfs.New(testfs.WithStat(func(string) (ihfs.FileInfo, error) {
				return nil, fs.ErrNotExist
			}))

			exists, err := ihfs.Exists(fsys, "nonexistent")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("should return error when stat returns an error", func() {
			testErr := errors.New("test error")
			fsys := testfs.New(testfs.WithStat(func(string) (ihfs.FileInfo, error) {
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
			fsys := testfs.New(testfs.WithStat(func(string) (ihfs.FileInfo, error) {
				return nil, fs.ErrNotExist
			}))

			isDir, err := ihfs.IsDir(fsys, "nonexistent")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fs.ErrNotExist))
			Expect(isDir).To(BeFalse())
		})

		It("should return error when stat returns an error", func() {
			testErr := errors.New("test error")
			fsys := testfs.New(testfs.WithStat(func(string) (ihfs.FileInfo, error) {
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
			fsys := testfs.New(testfs.WithWriteFile(func(string, []byte, ihfs.FileMode) error {
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
			fsys := testfs.New(testfs.WithWriteFile(func(string, []byte, ihfs.FileMode) error {
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
			fsys := testfs.New(testfs.WithMkdir(func(string, ihfs.FileMode) error {
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
				mkdirFunc: func(path string, _ ihfs.FileMode) error {
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
				mkdirFunc: func(path string, _ ihfs.FileMode) error {
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
				mkdirFunc: func(string, ihfs.FileMode) error {
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
				mkdirFunc: func(string, ihfs.FileMode) error {
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
				mkdirFunc: func(path string, _ ihfs.FileMode) error {
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
			fsys := testfs.New(testfs.WithMkdirAll(func(string, ihfs.FileMode) error {
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

func (r *errorReader) Read(_ []byte) (n int, err error) {
	return 0, r.err
}

type mkdirOnlyFS struct {
	mkdirFunc func(string, ihfs.FileMode) error
}

func (m *mkdirOnlyFS) Open(_ string) (ihfs.File, error) {
	return nil, fs.ErrNotExist
}

func (m *mkdirOnlyFS) Mkdir(name string, mode ihfs.FileMode) error {
	return m.mkdirFunc(name, mode)
}

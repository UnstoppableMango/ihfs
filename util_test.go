package ihfs_test

import (
	"bytes"
	"errors"
	"io"
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

type mockFileInfo struct {
	name  string
	isDir bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() fs.FileMode  { return 0 }
func (m mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() any           { return nil }

var _ = Describe("Util", func() {
	Describe("Copy", func() {
		var tempDir string

		BeforeEach(func() {
			tempDir = GinkgoT().TempDir()
		})

		It("should copy files from source filesystem to directory", func() {
			srcFs := osfs.New()
			destDir := filepath.Join(tempDir, "dest")

			err := ihfs.Copy(destDir, srcFs)

			Expect(err).NotTo(HaveOccurred())

			// Verify files were copied
			oneContent, err := os.ReadFile(filepath.Join(destDir, "testdata", "2-files", "one.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(oneContent)).To(Equal("one\n"))

			twoContent, err := os.ReadFile(filepath.Join(destDir, "testdata", "2-files", "two.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(twoContent)).To(Equal("two\n"))
		})

		It("should create destination directory if it does not exist", func() {
			srcFs := testfs.New(
				testfs.WithWalk(func(root string, fn fs.WalkDirFunc) error {
					entry := testfs.NewDirEntry("test.txt", false)
					return fn("test.txt", entry, nil)
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return testfs.NewFile(testfs.WithRead(func(p []byte) (int, error) {
						copy(p, []byte("content"))
						return len("content"), io.EOF
					})), nil
				}),
			)
			destDir := filepath.Join(tempDir, "new-dir")

			err := ihfs.Copy(destDir, srcFs)

			Expect(err).NotTo(HaveOccurred())
			_, err = os.Stat(destDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error if file already exists", func() {
			srcFs := testfs.New(
				testfs.WithWalk(func(root string, fn fs.WalkDirFunc) error {
					entry := testfs.NewDirEntry("test.txt", false)
					return fn("test.txt", entry, nil)
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return testfs.NewFile(testfs.WithRead(func(p []byte) (int, error) {
						copy(p, []byte("content"))
						return len("content"), io.EOF
					})), nil
				}),
			)

			// Create destination file first
			destDir := filepath.Join(tempDir, "dest")
			os.MkdirAll(destDir, 0755)
			os.WriteFile(filepath.Join(destDir, "test.txt"), []byte("existing"), 0644)

			err := ihfs.Copy(destDir, srcFs)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, fs.ErrExist)).To(BeTrue())
		})

		It("should handle directory creation", func() {
			srcFs := testfs.New(
				testfs.WithWalk(func(root string, fn fs.WalkDirFunc) error {
					dirEntry := testfs.NewDirEntry("subdir", true)
					if err := fn("subdir", dirEntry, nil); err != nil {
						return err
					}
					fileEntry := testfs.NewDirEntry("file.txt", false)
					return fn("file.txt", fileEntry, nil)
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return testfs.NewFile(testfs.WithRead(func(p []byte) (int, error) {
						return 0, io.EOF
					})), nil
				}),
			)
			destDir := filepath.Join(tempDir, "dest")

			err := ihfs.Copy(destDir, srcFs)

			Expect(err).NotTo(HaveOccurred())
			info, err := os.Stat(filepath.Join(destDir, "subdir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
		})

		It("should use CopyFS interface when available", func() {
			var capturedDir string
			var capturedSrc ihfs.FS
			copyFs := testfs.New(
				testfs.WithCopy(func(dir string, src ihfs.FS) error {
					capturedDir = dir
					capturedSrc = src
					return nil
				}),
			)

			err := ihfs.Copy("target", copyFs)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedDir).To(Equal("target"))
			Expect(capturedSrc).To(Equal(copyFs))
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

		It("should return error when Walk returns error", func() {
			walkErr := errors.New("walk failed")
			srcFs := testfs.New(
				testfs.WithWalk(func(root string, fn fs.WalkDirFunc) error {
					return walkErr
				}),
			)

			err := ihfs.Copy(tempDir, srcFs)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(walkErr))
		})

		It("should return error when Open returns error", func() {
			openErr := errors.New("open failed")
			srcFs := testfs.New(
				testfs.WithWalk(func(root string, fn fs.WalkDirFunc) error {
					entry := testfs.NewDirEntry("test.txt", false)
					return fn("test.txt", entry, nil)
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return nil, openErr
				}),
			)

			err := ihfs.Copy(tempDir, srcFs)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(openErr))
		})

		It("should clean up on copy error", func() {
			readErr := errors.New("read failed")
			srcFs := testfs.New(
				testfs.WithWalk(func(root string, fn fs.WalkDirFunc) error {
					entry := testfs.NewDirEntry("test.txt", false)
					return fn("test.txt", entry, nil)
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return testfs.NewFile(testfs.WithRead(func(p []byte) (int, error) {
						return 0, readErr
					})), nil
				}),
			)
			destDir := filepath.Join(tempDir, "dest")

			err := ihfs.Copy(destDir, srcFs)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(readErr))
			// Verify file was cleaned up
			_, err = os.Stat(filepath.Join(destDir, "test.txt"))
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})

	Describe("DirExists", func() {
		It("should return true when path is a directory", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return mockFileInfo{name: s, isDir: true}, nil
			}))

			exists, err := ihfs.DirExists(fsys, "dir")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("should return false when path is a file", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return mockFileInfo{name: s, isDir: false}, nil
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
				return mockFileInfo{name: s, isDir: false}, nil
			}))

			exists, err := ihfs.Exists(fsys, "file.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("should return true for directory", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return mockFileInfo{name: s, isDir: true}, nil
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
				return mockFileInfo{name: s, isDir: true}, nil
			}))

			isDir, err := ihfs.IsDir(fsys, "dir")

			Expect(err).NotTo(HaveOccurred())
			Expect(isDir).To(BeTrue())
		})

		It("should return false when path is a file", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return mockFileInfo{name: s, isDir: false}, nil
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

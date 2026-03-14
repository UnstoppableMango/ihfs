package ihfs_test

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/memfs"
	"github.com/unstoppablemango/ihfs/osfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("Util", func() {
	Describe("Copy", func() {
		It("should copy files from source filesystem to directory", func() {
			src, dest := memfs.New(), memfs.New()
			Expect(src.Mkdir("subdir", 0755)).NotTo(HaveOccurred())
			f, err := src.Create("subdir/test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			err = ihfs.Copy(dest, "", src)

			Expect(err).NotTo(HaveOccurred())
			dir, err := dest.Stat("subdir")
			Expect(err).NotTo(HaveOccurred())
			Expect(dir.IsDir()).To(BeTrue())
			file, err := dest.Stat("subdir/test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(file.IsDir()).To(BeFalse())
		})

		It("should delegate to CopyFS when dest implements CopyFS", func() {
			var capturedDir string

			src := testfs.New()
			dest := testfs.New(testfs.WithCopy(func(dir string, s ihfs.FS) error {
				capturedDir = dir
				return nil
			}))

			err := ihfs.Copy(dest, "mydir", src)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedDir).To(Equal("mydir"))
		})

		It("should propagate errors from Walk", func() {
			walkErr := errors.New("walk error")
			src := testfs.New(testfs.WithStat(func(string) (ihfs.FileInfo, error) {
				return nil, walkErr
			}))

			err := ihfs.Copy(testfs.BoringFs{}, "dir", src)

			Expect(err).To(MatchError(walkErr))
		})

		It("should propagate MkdirAll error for directory entries", func() {
			mkdirErr := errors.New("mkdir error")
			src := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
				fi := testfs.NewFileInfo(name)
				fi.IsDirFunc = func() bool { return true }
				fi.ModeFunc = func() fs.FileMode { return fs.ModeDir }
				return fi, nil
			}))
			dest := noSymlinkFS{mkdirAllFunc: func(string, ihfs.FileMode) error {
				return mkdirErr
			}}

			err := ihfs.Copy(dest, "dir", src)

			Expect(err).To(MatchError(mkdirErr))
		})

		It("should propagate Info error for directory entries", func() {
			infoErr := errors.New("info error")
			entry := testfs.NewDirEntry("subdir", true)
			entry.TypeFunc = func() ihfs.FileMode { return fs.ModeDir }
			entry.InfoFunc = func() (ihfs.FileInfo, error) { return nil, infoErr }
			src := testfs.New(
				withRootDirStat(),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
			)

			err := ihfs.Copy(noSymlinkFS{}, "dir", src)

			Expect(err).To(MatchError(infoErr))
		})
		It("should create symlink when dest implements SymlinkFS", func() {
			var capturedOld, capturedNew string

			entry := testfs.NewDirEntry("link.txt", false)
			entry.TypeFunc = func() ihfs.FileMode { return fs.ModeSymlink }
			src := testfs.New(
				withRootDirStat(),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
				testfs.WithReadLink(func(string) (string, error) {
					return "target", nil
				}),
			)
			dest := &copyDestFS{symlinkFunc: func(old, new string) error {
				capturedOld = old
				capturedNew = new
				return nil
			}}

			err := ihfs.Copy(dest, "dir", src)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedOld).To(Equal("target"))
			Expect(capturedNew).To(Equal("dir/link.txt"))
		})

		It("should propagate ReadLink error for symlinks", func() {
			readLinkErr := errors.New("readlink error")
			entry := testfs.NewDirEntry("link.txt", false)
			entry.TypeFunc = func() ihfs.FileMode { return fs.ModeSymlink }
			src := testfs.New(
				withRootDirStat(),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
				testfs.WithReadLink(func(string) (string, error) {
					return "", readLinkErr
				}),
			)

			err := ihfs.Copy(noSymlinkFS{}, "dir", src)

			Expect(err).To(MatchError(readLinkErr))
		})

		It("should return ErrNotImplemented when dest has no SymlinkFS for symlinks", func() {
			entry := testfs.NewDirEntry("link.txt", false)
			entry.TypeFunc = func() ihfs.FileMode { return fs.ModeSymlink }
			src := testfs.New(
				withRootDirStat(),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
				testfs.WithReadLink(func(string) (string, error) {
					return "target", nil
				}),
			)

			err := ihfs.Copy(noSymlinkFS{}, "dir", src)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
		})

		It("should return ErrNotImplemented when src doesn't implement ReadLinkFS for symlinks", func() {
			entry := testfs.NewDirEntry("link.txt", false)
			entry.TypeFunc = func() ihfs.FileMode { return fs.ModeSymlink }
			src := &noReadLinkFS{
				readDirFunc: func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				},
			}

			err := ihfs.Copy(noSymlinkFS{}, "dir", src)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
		})

		It("should propagate Open error for regular files", func() {
			openErr := errors.New("open error")
			entry := testfs.NewDirEntry("file.txt", false)
			src := testfs.New(
				withRootDirStat(),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return nil, openErr
				}),
			)

			err := ihfs.Copy(noSymlinkFS{}, "dir", src)

			Expect(err).To(MatchError(openErr))
		})

		It("should propagate Stat error for regular files", func() {
			statErr := errors.New("stat error")
			srcFile := &testfs.File{
				StatFunc:  func() (ihfs.FileInfo, error) { return nil, statErr },
				CloseFunc: func() error { return nil },
			}
			entry := testfs.NewDirEntry("file.txt", false)
			src := testfs.New(
				withRootDirStat(),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return srcFile, nil
				}),
			)

			err := ihfs.Copy(noSymlinkFS{}, "dir", src)

			Expect(err).To(MatchError(statErr))
		})

		It("should propagate OpenFile error for regular files", func() {
			openFileErr := errors.New("openfile error")
			srcFile := &testfs.File{
				StatFunc:  func() (ihfs.FileInfo, error) { return testfs.NewFileInfo("file.txt"), nil },
				CloseFunc: func() error { return nil },
			}
			entry := testfs.NewDirEntry("file.txt", false)
			src := testfs.New(
				withRootDirStat(),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return srcFile, nil
				}),
			)
			dest := &copyDestFS{openFileFunc: func(string, int, ihfs.FileMode) (ihfs.File, error) {
				return nil, openFileErr
			}}

			err := ihfs.Copy(dest, "dir", src)

			Expect(err).To(MatchError(openFileErr))
		})

		It("should return ErrNotImplemented when dest file is not io.Writer", func() {
			srcFile := &testfs.File{
				StatFunc:  func() (ihfs.FileInfo, error) { return testfs.NewFileInfo("file.txt"), nil },
				CloseFunc: func() error { return nil },
			}
			destFile := testfs.BoringFile{CloseFunc: func() error { return nil }}
			entry := testfs.NewDirEntry("file.txt", false)
			src := testfs.New(
				withRootDirStat(),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return srcFile, nil
				}),
			)
			dest := &copyDestFS{openFileFunc: func(string, int, ihfs.FileMode) (ihfs.File, error) {
				return destFile, nil
			}}

			err := ihfs.Copy(dest, "dir", src)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
		})

		It("should return PathError when io.Copy fails", func() {
			copyErr := errors.New("copy error")
			srcFile := &testfs.File{
				StatFunc:  func() (ihfs.FileInfo, error) { return testfs.NewFileInfo("file.txt"), nil },
				CloseFunc: func() error { return nil },
				ReadFunc:  func([]byte) (int, error) { return 0, copyErr },
			}
			destFile := &testfs.File{
				WriteFunc: func(p []byte) (int, error) { return len(p), nil },
				CloseFunc: func() error { return nil },
			}
			entry := testfs.NewDirEntry("file.txt", false)
			src := testfs.New(
				withRootDirStat(),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return srcFile, nil
				}),
			)
			dest := &copyDestFS{openFileFunc: func(string, int, ihfs.FileMode) (ihfs.File, error) {
				return destFile, nil
			}}

			err := ihfs.Copy(dest, "dir", src)

			Expect(err).To(HaveOccurred())
			var pathErr *fs.PathError
			Expect(errors.As(err, &pathErr)).To(BeTrue())
			Expect(pathErr.Op).To(Equal("Copy"))
			Expect(errors.Is(err, copyErr)).To(BeTrue())
		})

		It("should return error when dest file close fails", func() {
			closeErr := errors.New("close error")
			srcFile := &testfs.File{
				StatFunc:  func() (ihfs.FileInfo, error) { return testfs.NewFileInfo("file.txt"), nil },
				CloseFunc: func() error { return nil },
				ReadFunc:  func([]byte) (int, error) { return 0, io.EOF },
			}
			destFile := &testfs.File{
				WriteFunc: func(p []byte) (int, error) { return len(p), nil },
				CloseFunc: func() error { return closeErr },
			}
			entry := testfs.NewDirEntry("file.txt", false)
			src := testfs.New(
				withRootDirStat(),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return srcFile, nil
				}),
			)
			dest := &copyDestFS{openFileFunc: func(string, int, ihfs.FileMode) (ihfs.File, error) {
				return destFile, nil
			}}

			err := ihfs.Copy(dest, "dir", src)

			Expect(err).To(MatchError(closeErr))
		})

		It("should return PathError for unrecognized file types", func() {
			entry := testfs.NewDirEntry("device", false)
			entry.TypeFunc = func() ihfs.FileMode { return fs.ModeDevice }
			src := testfs.New(
				withRootDirStat(),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				}),
			)

			err := ihfs.Copy(noSymlinkFS{}, "dir", src)

			Expect(err).To(HaveOccurred())
			var pathErr *fs.PathError
			Expect(errors.As(err, &pathErr)).To(BeTrue())
			Expect(pathErr.Op).To(Equal("Copy"))
			Expect(errors.Is(err, fs.ErrInvalid)).To(BeTrue())
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

	Describe("OpenFile", func() {
		It("should call underlying OpenFile when OpenFileFS is implemented", func() {
			var capturedName string
			var capturedFlag int
			var capturedPerm ihfs.FileMode

			fsys := testfs.New(testfs.WithOpenFile(func(name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
				capturedName = name
				capturedFlag = flag
				capturedPerm = perm
				return &testfs.File{CloseFunc: func() error { return nil }}, nil
			}))

			f, err := ihfs.OpenFile(fsys, "test.txt", 0, 0o644)

			Expect(err).NotTo(HaveOccurred())
			Expect(f).NotTo(BeNil())
			Expect(capturedName).To(Equal("test.txt"))
			Expect(capturedFlag).To(Equal(0))
			Expect(capturedPerm).To(Equal(ihfs.FileMode(0o644)))
		})

		It("should return ErrNotImplemented when OpenFileFS not implemented", func() {
			f, err := ihfs.OpenFile(testfs.BoringFs{}, "test.txt", 0, 0o644)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
			Expect(f).To(BeNil())
		})
	})

	Describe("Remove", func() {
		It("should call underlying Remove when RemoveFS is implemented", func() {
			var capturedName string

			fsys := testfs.New(testfs.WithRemove(func(name string) error {
				capturedName = name
				return nil
			}))

			err := ihfs.Remove(fsys, "test.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("test.txt"))
		})

		It("should return ErrNotImplemented when RemoveFS not implemented", func() {
			err := ihfs.Remove(testfs.BoringFs{}, "test.txt")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
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

	Describe("Chmod", func() {
		It("should call underlying Chmod when ChmodFS is implemented", func() {
			var capturedName string
			var capturedMode ihfs.FileMode

			fsys := testfs.New(testfs.WithChmod(func(name string, mode ihfs.FileMode) error {
				capturedName = name
				capturedMode = mode
				return nil
			}))

			err := ihfs.Chmod(fsys, "file.txt", 0o644)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("file.txt"))
			Expect(capturedMode).To(Equal(ihfs.FileMode(0o644)))
		})

		It("should return ErrNotImplemented when ChmodFS not implemented", func() {
			err := ihfs.Chmod(testfs.BoringFs{}, "file.txt", 0o644)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
		})
	})

	Describe("Chown", func() {
		It("should call underlying Chown when ChownFS is implemented", func() {
			var capturedName string
			var capturedUID, capturedGID int

			fsys := testfs.New(testfs.WithChown(func(name string, uid, gid int) error {
				capturedName = name
				capturedUID = uid
				capturedGID = gid
				return nil
			}))

			err := ihfs.Chown(fsys, "file.txt", 1000, 1000)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("file.txt"))
			Expect(capturedUID).To(Equal(1000))
			Expect(capturedGID).To(Equal(1000))
		})

		It("should return ErrNotImplemented when ChownFS not implemented", func() {
			err := ihfs.Chown(testfs.BoringFs{}, "file.txt", 1000, 1000)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
		})
	})

	Describe("Chtimes", func() {
		It("should call underlying Chtimes when ChtimesFS is implemented", func() {
			var capturedName string
			var capturedAtime, capturedMtime time.Time

			fsys := testfs.New(testfs.WithChtimes(func(name string, atime, mtime time.Time) error {
				capturedName = name
				capturedAtime = atime
				capturedMtime = mtime
				return nil
			}))

			atime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
			mtime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
			err := ihfs.Chtimes(fsys, "file.txt", atime, mtime)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("file.txt"))
			Expect(capturedAtime).To(Equal(atime))
			Expect(capturedMtime).To(Equal(mtime))
		})

		It("should return ErrNotImplemented when ChtimesFS not implemented", func() {
			err := ihfs.Chtimes(testfs.BoringFs{}, "file.txt", time.Time{}, time.Time{})

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
		})
	})

	Describe("Create", func() {
		It("should call underlying Create when CreateFS is implemented", func() {
			var capturedName string
			expectedFile := &testfs.File{}

			fsys := testfs.New(testfs.WithCreate(func(name string) (ihfs.File, error) {
				capturedName = name
				return expectedFile, nil
			}))

			f, err := ihfs.Create(fsys, "file.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(f).To(Equal(expectedFile))
			Expect(capturedName).To(Equal("file.txt"))
		})

		It("should return ErrNotImplemented when CreateFS not implemented", func() {
			f, err := ihfs.Create(testfs.BoringFs{}, "file.txt")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
			Expect(f).To(BeNil())
		})
	})

	Describe("CreateTemp", func() {
		It("should call underlying CreateTemp when CreateTempFS is implemented", func() {
			var capturedDir, capturedPattern string
			expectedFile := &testfs.File{}

			fsys := testfs.New(testfs.WithCreateTemp(func(dir, pattern string) (ihfs.File, error) {
				capturedDir = dir
				capturedPattern = pattern
				return expectedFile, nil
			}))

			f, err := ihfs.CreateTemp(fsys, "/tmp", "prefix-*")

			Expect(err).NotTo(HaveOccurred())
			Expect(f).To(Equal(expectedFile))
			Expect(capturedDir).To(Equal("/tmp"))
			Expect(capturedPattern).To(Equal("prefix-*"))
		})

		It("should return ErrNotImplemented when CreateTempFS not implemented", func() {
			f, err := ihfs.CreateTemp(testfs.BoringFs{}, "/tmp", "prefix-*")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
			Expect(f).To(BeNil())
		})
	})

	Describe("MkdirTemp", func() {
		It("should call underlying MkdirTemp when MkdirTempFS is implemented", func() {
			var capturedDir, capturedPattern string

			fsys := testfs.New(testfs.WithMkdirTemp(func(dir, pattern string) (string, error) {
				capturedDir = dir
				capturedPattern = pattern
				return "/tmp/test123", nil
			}))

			name, err := ihfs.MkdirTemp(fsys, "/tmp", "test*")

			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("/tmp/test123"))
			Expect(capturedDir).To(Equal("/tmp"))
			Expect(capturedPattern).To(Equal("test*"))
		})

		It("should return ErrNotImplemented when MkdirTempFS not implemented", func() {
			name, err := ihfs.MkdirTemp(testfs.BoringFs{}, "/tmp", "test*")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
			Expect(name).To(BeEmpty())
		})
	})

	Describe("ReadFile", func() {
		It("should call underlying ReadFile when ReadFileFS is implemented", func() {
			var capturedName string

			fsys := testfs.New(testfs.WithReadFile(func(name string) ([]byte, error) {
				capturedName = name
				return []byte("content"), nil
			}))

			data, err := ihfs.ReadFile(fsys, "file.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal([]byte("content")))
			Expect(capturedName).To(Equal("file.txt"))
		})

		It("should return ErrNotImplemented when ReadFileFS not implemented", func() {
			data, err := ihfs.ReadFile(testfs.BoringFs{}, "file.txt")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
			Expect(data).To(BeNil())
		})
	})

	Describe("ReadLink", func() {
		It("should call underlying ReadLink when ReadLinkFS is implemented", func() {
			var capturedName string

			fsys := testfs.New(testfs.WithReadLink(func(name string) (string, error) {
				capturedName = name
				return "target", nil
			}))

			target, err := ihfs.ReadLink(fsys, "symlink")

			Expect(err).NotTo(HaveOccurred())
			Expect(target).To(Equal("target"))
			Expect(capturedName).To(Equal("symlink"))
		})

		It("should return ErrNotImplemented when ReadLinkFS not implemented", func() {
			target, err := ihfs.ReadLink(testfs.BoringFs{}, "symlink")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
			Expect(target).To(BeEmpty())
		})
	})

	Describe("RemoveAll", func() {
		It("should call underlying RemoveAll when RemoveAllFS is implemented", func() {
			var capturedName string

			fsys := testfs.New(testfs.WithRemoveAll(func(name string) error {
				capturedName = name
				return nil
			}))

			err := ihfs.RemoveAll(fsys, "dir")

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("dir"))
		})

		It("should return ErrNotImplemented when RemoveAllFS not implemented", func() {
			err := ihfs.RemoveAll(testfs.BoringFs{}, "dir")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
		})
	})

	Describe("Rename", func() {
		It("should call underlying Rename when RenameFS is implemented", func() {
			var capturedOldpath, capturedNewpath string

			fsys := testfs.New(testfs.WithRename(func(oldpath, newpath string) error {
				capturedOldpath = oldpath
				capturedNewpath = newpath
				return nil
			}))

			err := ihfs.Rename(fsys, "old.txt", "new.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedOldpath).To(Equal("old.txt"))
			Expect(capturedNewpath).To(Equal("new.txt"))
		})

		It("should return ErrNotImplemented when RenameFS not implemented", func() {
			err := ihfs.Rename(testfs.BoringFs{}, "old.txt", "new.txt")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
		})
	})

	Describe("Sub", func() {
		It("should call underlying Sub when SubFS is implemented", func() {
			var capturedDir string
			expectedFS := testfs.New()

			fsys := testfs.New(testfs.WithSub(func(dir string) (ihfs.FS, error) {
				capturedDir = dir
				return &expectedFS, nil
			}))

			subFS, err := ihfs.Sub(fsys, "subdir")

			Expect(err).NotTo(HaveOccurred())
			Expect(subFS).To(BeIdenticalTo(&expectedFS))
			Expect(capturedDir).To(Equal("subdir"))
		})

		It("should return ErrNotImplemented when SubFS not implemented", func() {
			subFS, err := ihfs.Sub(testfs.BoringFs{}, "subdir")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
			Expect(subFS).To(BeNil())
		})
	})

	Describe("Symlink", func() {
		It("should call underlying Symlink when SymlinkFS is implemented", func() {
			var capturedOldname, capturedNewname string

			fsys := testfs.New(testfs.WithSymlink(func(oldname, newname string) error {
				capturedOldname = oldname
				capturedNewname = newname
				return nil
			}))

			err := ihfs.Symlink(fsys, "target", "link")

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedOldname).To(Equal("target"))
			Expect(capturedNewname).To(Equal("link"))
		})

		It("should return ErrNotImplemented when SymlinkFS not implemented", func() {
			err := ihfs.Symlink(testfs.BoringFs{}, "target", "link")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
		})
	})

	Describe("TempFile", func() {
		It("should call underlying TempFile when TempFileFS is implemented", func() {
			var capturedDir, capturedPattern string

			fsys := testfs.New(testfs.WithTempFile(func(dir, pattern string) (string, error) {
				capturedDir = dir
				capturedPattern = pattern
				return "/tmp/tempfile123", nil
			}))

			name, err := ihfs.TempFile(fsys, "/tmp", "prefix-*")

			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("/tmp/tempfile123"))
			Expect(capturedDir).To(Equal("/tmp"))
			Expect(capturedPattern).To(Equal("prefix-*"))
		})

		It("should return ErrNotImplemented when TempFileFS not implemented", func() {
			name, err := ihfs.TempFile(testfs.BoringFs{}, "/tmp", "prefix-*")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrNotImplemented)).To(BeTrue())
			Expect(name).To(BeEmpty())
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

// noSymlinkFS implements MkdirAllFS but not SymlinkFS or CopyFS.
type noSymlinkFS struct {
	mkdirAllFunc func(string, ihfs.FileMode) error
}

func (noSymlinkFS) Open(string) (ihfs.File, error) {
	return nil, fs.ErrNotExist
}

func (f noSymlinkFS) MkdirAll(name string, perm ihfs.FileMode) error {
	if f.mkdirAllFunc != nil {
		return f.mkdirAllFunc(name, perm)
	}
	return nil
}

// copyDestFS is a configurable test filesystem for the Copy fallback path.
// It intentionally does NOT implement CopyFS.
type copyDestFS struct {
	mkdirAllFunc func(string, ihfs.FileMode) error
	openFileFunc func(string, int, ihfs.FileMode) (ihfs.File, error)
	symlinkFunc  func(string, string) error
}

func (*copyDestFS) Open(string) (ihfs.File, error) {
	return nil, fs.ErrNotExist
}

func (d *copyDestFS) MkdirAll(name string, perm ihfs.FileMode) error {
	if d.mkdirAllFunc != nil {
		return d.mkdirAllFunc(name, perm)
	}
	return nil
}

func (d *copyDestFS) OpenFile(name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
	if d.openFileFunc != nil {
		return d.openFileFunc(name, flag, perm)
	}
	return nil, errors.New("openfile: not implemented")
}

func (d *copyDestFS) Symlink(old, new string) error {
	if d.symlinkFunc != nil {
		return d.symlinkFunc(old, new)
	}
	return nil
}

// noReadLinkFS implements a minimal FS with ReadDir but without ReadLinkFS.
type noReadLinkFS struct {
	readDirFunc func(string) ([]ihfs.DirEntry, error)
}

func (*noReadLinkFS) Open(string) (ihfs.File, error) {
	return nil, fs.ErrNotExist
}

func (f *noReadLinkFS) Stat(name string) (ihfs.FileInfo, error) {
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

func (f *noReadLinkFS) ReadDir(name string) ([]ihfs.DirEntry, error) {
	return f.readDirFunc(name)
}

// withRootDirStat returns a testfs.Option that configures Stat to report "." as a directory.
func withRootDirStat() testfs.Option {
	return testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
		fi := testfs.NewFileInfo(name)
		fi.IsDirFunc = func() bool { return name == "." }
		fi.ModeFunc = func() fs.FileMode {
			if name == "." {
				return fs.ModeDir
			}
			return 0
		}
		return fi, nil
	})
}

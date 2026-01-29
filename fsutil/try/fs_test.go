package try_test

import (
	"errors"
	"io/fs"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/fsutil/try"
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

var _ = Describe("Try Util", func() {
	Describe("DirExists", func() {
		It("should return true for directory", func() {
			fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
				return mockFileInfo{name: name, isDir: true}, nil
			}))

			exists, err := try.DirExists(fsys, "dir")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("should return false for file", func() {
			fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
				return mockFileInfo{name: name, isDir: false}, nil
			}))

			exists, err := try.DirExists(fsys, "file.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("should return false for nonexistent path", func() {
			fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
				return nil, fs.ErrNotExist
			}))

			exists, err := try.DirExists(fsys, "nonexistent")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("should return ErrUnsupported for non-Stat filesystem", func() {
			fsys := testfs.BoringFs{}

			exists, err := try.DirExists(fsys, "dir")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
			Expect(exists).To(BeFalse())
		})
	})

	Describe("Exists", func() {
		It("should return true for file", func() {
			fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
				return mockFileInfo{name: name, isDir: false}, nil
			}))

			exists, err := try.Exists(fsys, "file.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("should return true for directory", func() {
			fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
				return mockFileInfo{name: name, isDir: true}, nil
			}))

			exists, err := try.Exists(fsys, "dir")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("should return false for nonexistent path", func() {
			fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
				return nil, fs.ErrNotExist
			}))

			exists, err := try.Exists(fsys, "nonexistent")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("should return ErrUnsupported for non-Stat filesystem", func() {
			fsys := testfs.BoringFs{}

			exists, err := try.Exists(fsys, "file.txt")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
			Expect(exists).To(BeFalse())
		})
	})

	Describe("Stat", func() {
		It("should return FileInfo for file", func() {
			fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
				return mockFileInfo{name: name, isDir: false}, nil
			}))

			info, err := try.Stat(fsys, "file.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Name()).To(Equal("file.txt"))
		})

		It("should return FileInfo for directory", func() {
			fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
				return mockFileInfo{name: name, isDir: true}, nil
			}))

			info, err := try.Stat(fsys, "dir")

			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Name()).To(Equal("dir"))
			Expect(info.IsDir()).To(BeTrue())
		})

		It("should return error for nonexistent path", func() {
			fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
				return nil, fs.ErrNotExist
			}))

			info, err := try.Stat(fsys, "nonexistent")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, fs.ErrNotExist)).To(BeTrue())
			Expect(info).To(BeNil())
		})

		It("should return ErrUnsupported for non-Stat filesystem", func() {
			fsys := testfs.BoringFs{}

			info, err := try.Stat(fsys, "file.txt")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
			Expect(info).To(BeNil())
		})
	})

	Describe("IsDir", func() {
		It("should return true for directory", func() {
			fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
				return mockFileInfo{name: name, isDir: true}, nil
			}))

			isDir, err := try.IsDir(fsys, "dir")

			Expect(err).NotTo(HaveOccurred())
			Expect(isDir).To(BeTrue())
		})

		It("should return false for file", func() {
			fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
				return mockFileInfo{name: name, isDir: false}, nil
			}))

			isDir, err := try.IsDir(fsys, "file.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(isDir).To(BeFalse())
		})

		It("should return error for nonexistent path", func() {
			fsys := testfs.New(testfs.WithStat(func(s string) (ihfs.FileInfo, error) {
				return nil, fs.ErrNotExist
			}))

			isDir, err := try.IsDir(fsys, "nonexistent")

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, fs.ErrNotExist)).To(BeTrue())
			Expect(isDir).To(BeFalse())
		})

		It("should return ErrUnsupported for non-Stat filesystem", func() {
			fsys := testfs.BoringFs{}

			isDir, err := try.IsDir(fsys, "dir")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
			Expect(isDir).To(BeFalse())
		})
	})

	Describe("ReadDir", func() {
		It("should read directory entries", func() {
			fsys := osfs.New()

			entries, err := try.ReadDir(fsys, "../../testdata/2-files")

			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(2))
			Expect(entries[0].Name()).To(Equal("one.txt"))
			Expect(entries[1].Name()).To(Equal("two.txt"))
		})

		It("should return error when fs does not support ReadDir", func() {
			fsys := testfs.BoringFs{}

			entries, err := try.ReadDir(fsys, "./nonexistent")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
			Expect(entries).To(BeNil())
		})
	})

	Describe("ReadDirNames", func() {
		It("should read directory entry names", func() {
			fsys := osfs.New()

			names, err := try.ReadDirNames(fsys, "../../testdata/2-files")

			Expect(err).NotTo(HaveOccurred())
			Expect(names).To(ConsistOf("one.txt", "two.txt"))
		})

		It("should return error when fs does not support ReadDir", func() {
			fsys := testfs.BoringFs{}

			names, err := try.ReadDirNames(fsys, "./nonexistent")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
			Expect(names).To(BeNil())
		})
	})

	Describe("Chmod", func() {
		It("should call Chmod on the filesystem", func() {
			var capturedName string
			var capturedMode ihfs.FileMode

			fsys := testfs.New(testfs.WithChmod(func(name string, mode ihfs.FileMode) error {
				capturedName = name
				capturedMode = mode
				return nil
			}))

			err := try.Chmod(fsys, "file.txt", 0o644)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("file.txt"))
			Expect(capturedMode).To(Equal(ihfs.FileMode(0o644)))
		})

		It("should return ErrNotSupported when fs does not support Chmod", func() {
			fsys := testfs.BoringFs{}

			err := try.Chmod(fsys, "file.txt", 0o644)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})
	})

	Describe("Chown", func() {
		It("should call Chown on the filesystem", func() {
			var capturedName string
			var capturedUid, capturedGid int

			fsys := testfs.New(testfs.WithChown(func(name string, uid, gid int) error {
				capturedName = name
				capturedUid = uid
				capturedGid = gid
				return nil
			}))

			err := try.Chown(fsys, "file.txt", 1000, 1000)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("file.txt"))
			Expect(capturedUid).To(Equal(1000))
			Expect(capturedGid).To(Equal(1000))
		})

		It("should return ErrNotSupported when fs does not support Chown", func() {
			fsys := testfs.BoringFs{}

			err := try.Chown(fsys, "file.txt", 1000, 1000)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})
	})

	Describe("Chtimes", func() {
		It("should call Chtimes on the filesystem", func() {
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
			err := try.Chtimes(fsys, "file.txt", atime, mtime)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("file.txt"))
			Expect(capturedAtime).To(Equal(atime))
			Expect(capturedMtime).To(Equal(mtime))
		})

		It("should return ErrNotSupported when fs does not support Chtimes", func() {
			fsys := testfs.BoringFs{}

			err := try.Chtimes(fsys, "file.txt", time.Time{}, time.Time{})

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})
	})

	Describe("Copy", func() {
		It("should call Copy on the filesystem", func() {
			var capturedDir string
			var capturedSrc ihfs.FS

			fsys := testfs.New(testfs.WithCopy(func(dir string, src ihfs.FS) error {
				capturedDir = dir
				capturedSrc = src
				return nil
			}))

			srcFs := osfs.New()
			err := try.Copy(fsys, "dest", srcFs)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedDir).To(Equal("dest"))
			Expect(capturedSrc).To(Equal(srcFs))
		})

		It("should return ErrNotSupported when fs does not support Copy", func() {
			fsys := testfs.BoringFs{}

			err := try.Copy(fsys, "dest", osfs.New())

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})
	})

	Describe("Mkdir", func() {
		It("should call Mkdir on the filesystem", func() {
			var capturedName string
			var capturedMode ihfs.FileMode

			fsys := testfs.New(testfs.WithMkdir(func(name string, mode ihfs.FileMode) error {
				capturedName = name
				capturedMode = mode
				return nil
			}))

			err := try.Mkdir(fsys, "newdir", 0o755)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("newdir"))
			Expect(capturedMode).To(Equal(ihfs.FileMode(0o755)))
		})

		It("should return ErrNotSupported when fs does not support Mkdir", func() {
			fsys := testfs.BoringFs{}

			err := try.Mkdir(fsys, "newdir", 0o755)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})
	})

	Describe("MkdirAll", func() {
		It("should call MkdirAll on the filesystem", func() {
			var capturedName string
			var capturedMode ihfs.FileMode

			fsys := testfs.New(testfs.WithMkdirAll(func(name string, mode ihfs.FileMode) error {
				capturedName = name
				capturedMode = mode
				return nil
			}))

			err := try.MkdirAll(fsys, "path/to/dir", 0o755)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("path/to/dir"))
			Expect(capturedMode).To(Equal(ihfs.FileMode(0o755)))
		})

		It("should return ErrNotSupported when fs does not support MkdirAll", func() {
			fsys := testfs.BoringFs{}

			err := try.MkdirAll(fsys, "path/to/dir", 0o755)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})
	})

	Describe("MkdirTemp", func() {
		It("should call MkdirTemp on the filesystem", func() {
			var capturedDir, capturedPattern string

			fsys := testfs.New(testfs.WithMkdirTemp(func(dir, pattern string) (string, error) {
				capturedDir = dir
				capturedPattern = pattern
				return "/tmp/test123", nil
			}))

			name, err := try.MkdirTemp(fsys, "/tmp", "test*")

			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("/tmp/test123"))
			Expect(capturedDir).To(Equal("/tmp"))
			Expect(capturedPattern).To(Equal("test*"))
		})

		It("should return ErrNotSupported when fs does not support MkdirTemp", func() {
			fsys := testfs.BoringFs{}

			name, err := try.MkdirTemp(fsys, "/tmp", "test*")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
			Expect(name).To(BeEmpty())
		})
	})

	Describe("Remove", func() {
		It("should call Remove on the filesystem", func() {
			var capturedName string

			fsys := testfs.New(testfs.WithRemove(func(name string) error {
				capturedName = name
				return nil
			}))

			err := try.Remove(fsys, "file.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("file.txt"))
		})

		It("should return ErrNotSupported when fs does not support Remove", func() {
			fsys := testfs.BoringFs{}

			err := try.Remove(fsys, "file.txt")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})
	})

	Describe("RemoveAll", func() {
		It("should call RemoveAll on the filesystem", func() {
			var capturedName string

			fsys := testfs.New(testfs.WithRemoveAll(func(name string) error {
				capturedName = name
				return nil
			}))

			err := try.RemoveAll(fsys, "dir")

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("dir"))
		})

		It("should return ErrNotSupported when fs does not support RemoveAll", func() {
			fsys := testfs.BoringFs{}

			err := try.RemoveAll(fsys, "dir")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})
	})

	Describe("WriteFile", func() {
		It("should call WriteFile on the filesystem", func() {
			var capturedName string
			var capturedData []byte
			var capturedPerm ihfs.FileMode

			fsys := testfs.New(testfs.WithWriteFile(func(name string, data []byte, perm ihfs.FileMode) error {
				capturedName = name
				capturedData = data
				capturedPerm = perm
				return nil
			}))

			err := try.WriteFile(fsys, "file.txt", []byte("content"), 0o644)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedName).To(Equal("file.txt"))
			Expect(capturedData).To(Equal([]byte("content")))
			Expect(capturedPerm).To(Equal(ihfs.FileMode(0o644)))
		})

		It("should return ErrNotSupported when fs does not support WriteFile", func() {
			fsys := testfs.BoringFs{}

			err := try.WriteFile(fsys, "file.txt", []byte("content"), 0o644)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})
	})
})

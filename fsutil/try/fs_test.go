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

type boringFS struct{ ihfs.FS }

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
			fsys := boringFS{FS: testfs.New()}

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
			fsys := boringFS{FS: testfs.New()}

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
			fsys := boringFS{FS: testfs.New()}

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
			fsys := boringFS{FS: testfs.New()}

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
			fsys := boringFS{FS: testfs.New()}

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
			fsys := boringFS{FS: testfs.New()}

			names, err := try.ReadDirNames(fsys, "./nonexistent")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
			Expect(names).To(BeNil())
		})
	})
})

package testfs_test

import (
	"errors"
	"io/fs"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("BoringFs", func() {
	Describe("Open", func() {
		It("calls OpenFunc when set", func() {
			called := false
			bf := testfs.BoringFs{
				OpenFunc: func(string) (ihfs.File, error) {
					called = true
					return nil, nil
				},
			}

			_, _ = bf.Open("test.txt")

			Expect(called).To(BeTrue())
		})

		It("returns ErrNotImplemented when OpenFunc is nil", func() {
			bf := testfs.BoringFs{}

			_, err := bf.Open("test.txt")

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})
	})
})

var _ = Describe("BoringFile", func() {
	Describe("Close", func() {
		It("calls CloseFunc when set", func() {
			called := false
			bf := testfs.BoringFile{CloseFunc: func() error { called = true; return nil }}

			_ = bf.Close()

			Expect(called).To(BeTrue())
		})

		It("returns ErrNotImplemented when CloseFunc is nil", func() {
			bf := testfs.BoringFile{}

			err := bf.Close()

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})
	})

	Describe("Read", func() {
		It("calls ReadFunc when set", func() {
			called := false
			bf := testfs.BoringFile{ReadFunc: func([]byte) (int, error) { called = true; return 0, nil }}

			_, _ = bf.Read(nil)

			Expect(called).To(BeTrue())
		})

		It("returns ErrNotImplemented when ReadFunc is nil", func() {
			bf := testfs.BoringFile{}

			_, err := bf.Read(nil)

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})
	})

	Describe("Stat", func() {
		It("calls StatFunc when set", func() {
			called := false
			bf := testfs.BoringFile{StatFunc: func() (ihfs.FileInfo, error) { called = true; return nil, nil }}

			_, _ = bf.Stat()

			Expect(called).To(BeTrue())
		})

		It("returns ErrNotImplemented when StatFunc is nil", func() {
			bf := testfs.BoringFile{}

			_, err := bf.Stat()

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})
	})
})

var _ = Describe("File", func() {
	Describe("Name", func() {
		It("returns empty string for zero value", func() {
			f := &testfs.File{}

			Expect(f.Name()).To(Equal(""))
		})
	})

	Describe("Close", func() {
		It("returns nil when CloseFunc is nil", func() {
			f := &testfs.File{}

			Expect(f.Close()).To(Succeed())
		})

		It("calls CloseFunc when set", func() {
			called := false
			f := &testfs.File{CloseFunc: func() error { called = true; return nil }}

			_ = f.Close()

			Expect(called).To(BeTrue())
		})
	})

	Describe("Read", func() {
		It("returns ErrNotImplemented when ReadFunc is nil", func() {
			f := &testfs.File{}

			_, err := f.Read(nil)

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})

		It("calls ReadFunc when set", func() {
			called := false
			f := &testfs.File{ReadFunc: func([]byte) (int, error) { called = true; return 0, nil }}

			_, _ = f.Read(nil)

			Expect(called).To(BeTrue())
		})
	})

	Describe("Stat", func() {
		It("returns ErrNotImplemented when StatFunc is nil", func() {
			f := &testfs.File{}

			_, err := f.Stat()

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})

		It("calls StatFunc when set", func() {
			called := false
			f := &testfs.File{StatFunc: func() (ihfs.FileInfo, error) { called = true; return nil, nil }}

			_, _ = f.Stat()

			Expect(called).To(BeTrue())
		})
	})

	Describe("Seek", func() {
		It("returns ErrNotImplemented when SeekFunc is nil", func() {
			f := &testfs.File{}

			_, err := f.Seek(0, 0)

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})

		It("calls SeekFunc when set", func() {
			called := false
			f := &testfs.File{SeekFunc: func(int64, int) (int64, error) { called = true; return 0, nil }}

			_, _ = f.Seek(0, 0)

			Expect(called).To(BeTrue())
		})
	})

	Describe("Write", func() {
		It("returns ErrNotImplemented when WriteFunc is nil", func() {
			f := &testfs.File{}

			_, err := f.Write(nil)

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})

		It("calls WriteFunc when set", func() {
			called := false
			f := &testfs.File{WriteFunc: func([]byte) (int, error) { called = true; return 0, nil }}

			_, _ = f.Write(nil)

			Expect(called).To(BeTrue())
		})
	})

	Describe("ReadAt", func() {
		It("returns ErrNotImplemented when ReadAtFunc is nil", func() {
			f := &testfs.File{}

			_, err := f.ReadAt(nil, 0)

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})

		It("calls ReadAtFunc when set", func() {
			called := false
			f := &testfs.File{ReadAtFunc: func([]byte, int64) (int, error) { called = true; return 0, nil }}

			_, _ = f.ReadAt(nil, 0)

			Expect(called).To(BeTrue())
		})
	})

	Describe("WriteAt", func() {
		It("returns ErrNotImplemented when WriteAtFunc is nil", func() {
			f := &testfs.File{}

			_, err := f.WriteAt(nil, 0)

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})

		It("calls WriteAtFunc when set", func() {
			called := false
			f := &testfs.File{WriteAtFunc: func([]byte, int64) (int, error) { called = true; return 0, nil }}

			_, _ = f.WriteAt(nil, 0)

			Expect(called).To(BeTrue())
		})
	})

	Describe("WriteString", func() {
		It("returns ErrNotImplemented when WriteStringFunc is nil", func() {
			f := &testfs.File{}

			_, err := f.WriteString("x")

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})

		It("calls WriteStringFunc when set", func() {
			called := false
			f := &testfs.File{WriteStringFunc: func(string) (int, error) { called = true; return 0, nil }}

			_, _ = f.WriteString("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("Sync", func() {
		It("returns ErrNotImplemented when SyncFunc is nil", func() {
			f := &testfs.File{}

			err := f.Sync()

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})

		It("calls SyncFunc when set", func() {
			called := false
			f := &testfs.File{SyncFunc: func() error { called = true; return nil }}

			_ = f.Sync()

			Expect(called).To(BeTrue())
		})
	})

	Describe("Truncate", func() {
		It("returns ErrNotImplemented when TruncateFunc is nil", func() {
			f := &testfs.File{}

			err := f.Truncate(0)

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})

		It("calls TruncateFunc when set", func() {
			called := false
			f := &testfs.File{TruncateFunc: func(int64) error { called = true; return nil }}

			_ = f.Truncate(0)

			Expect(called).To(BeTrue())
		})
	})

	Describe("ReadDir", func() {
		It("returns ErrNotImplemented when ReadDirFunc is nil", func() {
			f := &testfs.File{}

			_, err := f.ReadDir(-1)

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})

		It("calls ReadDirFunc when set", func() {
			called := false
			f := &testfs.File{ReadDirFunc: func(int) ([]ihfs.DirEntry, error) { called = true; return nil, nil }}

			_, _ = f.ReadDir(-1)

			Expect(called).To(BeTrue())
		})
	})

	Describe("ReadDirNames", func() {
		It("returns ErrNotImplemented when ReadDirNamesFunc is nil", func() {
			f := &testfs.File{}

			_, err := f.ReadDirNames(-1)

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})

		It("calls ReadDirNamesFunc when set", func() {
			called := false
			f := &testfs.File{ReadDirNamesFunc: func(int) ([]string, error) { called = true; return nil, nil }}

			_, _ = f.ReadDirNames(-1)

			Expect(called).To(BeTrue())
		})
	})
})

var _ = Describe("DirEntry", func() {
	Describe("NewDirEntry", func() {
		It("creates a DirEntry with the given name and isDir", func() {
			d := testfs.NewDirEntry("mydir", true)

			Expect(d.Name()).To(Equal("mydir"))
			Expect(d.IsDir()).To(BeTrue())
			Expect(d.Type()).To(Equal(ihfs.FileMode(0)))
		})

		It("creates a non-directory DirEntry", func() {
			d := testfs.NewDirEntry("file.txt", false)

			Expect(d.Name()).To(Equal("file.txt"))
			Expect(d.IsDir()).To(BeFalse())
		})

		It("Info returns FileInfo with correct IsDir", func() {
			d := testfs.NewDirEntry("mydir", true)

			info, err := d.Info()

			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Name()).To(Equal("mydir"))
			Expect(info.IsDir()).To(BeTrue())
		})
	})

	Describe("IsDir", func() {
		It("calls IsDirFunc", func() {
			called := false
			d := &testfs.DirEntry{
				IsDirFunc: func() bool { called = true; return true },
				TypeFunc:  func() ihfs.FileMode { return 0 },
			}

			_ = d.IsDir()

			Expect(called).To(BeTrue())
		})
	})

	Describe("Type", func() {
		It("calls TypeFunc", func() {
			called := false
			d := &testfs.DirEntry{
				IsDirFunc: func() bool { return false },
				TypeFunc:  func() ihfs.FileMode { called = true; return 0 },
			}

			_ = d.Type()

			Expect(called).To(BeTrue())
		})
	})

	Describe("Info", func() {
		It("calls InfoFunc when set", func() {
			called := false
			d := &testfs.DirEntry{
				IsDirFunc: func() bool { return false },
				TypeFunc:  func() ihfs.FileMode { return 0 },
				InfoFunc:  func() (ihfs.FileInfo, error) { called = true; return nil, nil },
			}

			_, _ = d.Info()

			Expect(called).To(BeTrue())
		})

		It("returns ErrNotImplemented when InfoFunc is nil", func() {
			d := &testfs.DirEntry{}

			_, err := d.Info()

			Expect(errors.Is(err, testfs.ErrNotImplemented)).To(BeTrue())
		})
	})
})

var _ = Describe("FileInfo", func() {
	Describe("NewFileInfo", func() {
		It("creates FileInfo with the given name", func() {
			fi := testfs.NewFileInfo("test.txt")

			Expect(fi.Name()).To(Equal("test.txt"))
		})

		It("IsDir returns false by default", func() {
			fi := testfs.NewFileInfo("test.txt")

			Expect(fi.IsDir()).To(BeFalse())
		})

		It("Size returns 0 by default", func() {
			fi := testfs.NewFileInfo("test.txt")

			Expect(fi.Size()).To(BeZero())
		})

		It("Mode returns 0 by default", func() {
			fi := testfs.NewFileInfo("test.txt")

			Expect(fi.Mode()).To(Equal(fs.FileMode(0)))
		})

		It("ModTime returns zero time by default", func() {
			fi := testfs.NewFileInfo("test.txt")

			Expect(fi.ModTime()).To(Equal(time.Time{}))
		})

		It("Sys returns nil by default", func() {
			fi := testfs.NewFileInfo("test.txt")

			Expect(fi.Sys()).To(BeNil())
		})
	})
})

var _ = Describe("Fs", func() {
	Describe("Open", func() {
		It("returns ErrNotExist by default", func() {
			_, err := testfs.New().Open("x")

			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("calls custom function when set via WithOpen", func() {
			called := false
			f := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
				called = true
				return nil, nil
			}))

			_, _ = f.Open("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("Stat", func() {
		It("returns ErrNotExist by default", func() {
			_, err := testfs.New().Stat("x")

			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("calls custom function when set via WithStat", func() {
			called := false
			f := testfs.New(testfs.WithStat(func(string) (ihfs.FileInfo, error) {
				called = true
				return nil, nil
			}))

			_, _ = f.Stat("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("Create", func() {
		It("returns ErrPermission by default", func() {
			_, err := testfs.New().Create("x")

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithCreate", func() {
			called := false
			f := testfs.New(testfs.WithCreate(func(string) (ihfs.File, error) {
				called = true
				return nil, nil
			}))

			_, _ = f.Create("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("WriteFile", func() {
		It("returns ErrPermission by default", func() {
			err := testfs.New().WriteFile("x", nil, 0)

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithWriteFile", func() {
			called := false
			f := testfs.New(testfs.WithWriteFile(func(string, []byte, ihfs.FileMode) error {
				called = true
				return nil
			}))

			_ = f.WriteFile("x", nil, 0)

			Expect(called).To(BeTrue())
		})
	})

	Describe("Chmod", func() {
		It("returns ErrPermission by default", func() {
			err := testfs.New().Chmod("x", 0)

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithChmod", func() {
			called := false
			f := testfs.New(testfs.WithChmod(func(string, ihfs.FileMode) error {
				called = true
				return nil
			}))

			_ = f.Chmod("x", 0)

			Expect(called).To(BeTrue())
		})
	})

	Describe("Chown", func() {
		It("returns ErrPermission by default", func() {
			err := testfs.New().Chown("x", 0, 0)

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithChown", func() {
			called := false
			f := testfs.New(testfs.WithChown(func(string, int, int) error {
				called = true
				return nil
			}))

			_ = f.Chown("x", 0, 0)

			Expect(called).To(BeTrue())
		})
	})

	Describe("Chtimes", func() {
		It("returns ErrPermission by default", func() {
			err := testfs.New().Chtimes("x", time.Time{}, time.Time{})

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithChtimes", func() {
			called := false
			f := testfs.New(testfs.WithChtimes(func(string, time.Time, time.Time) error {
				called = true
				return nil
			}))

			_ = f.Chtimes("x", time.Time{}, time.Time{})

			Expect(called).To(BeTrue())
		})
	})

	Describe("Copy", func() {
		It("returns ErrPermission by default", func() {
			err := testfs.New().Copy("x", nil)

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithCopy", func() {
			called := false
			f := testfs.New(testfs.WithCopy(func(string, ihfs.FS) error {
				called = true
				return nil
			}))

			_ = f.Copy("x", nil)

			Expect(called).To(BeTrue())
		})
	})

	Describe("Mkdir", func() {
		It("returns ErrPermission by default", func() {
			err := testfs.New().Mkdir("x", 0)

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithMkdir", func() {
			called := false
			f := testfs.New(testfs.WithMkdir(func(string, ihfs.FileMode) error {
				called = true
				return nil
			}))

			_ = f.Mkdir("x", 0)

			Expect(called).To(BeTrue())
		})
	})

	Describe("MkdirAll", func() {
		It("returns ErrPermission by default", func() {
			err := testfs.New().MkdirAll("x", 0)

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithMkdirAll", func() {
			called := false
			f := testfs.New(testfs.WithMkdirAll(func(string, ihfs.FileMode) error {
				called = true
				return nil
			}))

			_ = f.MkdirAll("x", 0)

			Expect(called).To(BeTrue())
		})
	})

	Describe("MkdirTemp", func() {
		It("returns ErrPermission by default", func() {
			_, err := testfs.New().MkdirTemp("x", "y")

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithMkdirTemp", func() {
			called := false
			f := testfs.New(testfs.WithMkdirTemp(func(string, string) (string, error) {
				called = true
				return "", nil
			}))

			_, _ = f.MkdirTemp("x", "y")

			Expect(called).To(BeTrue())
		})
	})

	Describe("Remove", func() {
		It("returns ErrPermission by default", func() {
			err := testfs.New().Remove("x")

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithRemove", func() {
			called := false
			f := testfs.New(testfs.WithRemove(func(string) error {
				called = true
				return nil
			}))

			_ = f.Remove("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("RemoveAll", func() {
		It("returns ErrPermission by default", func() {
			err := testfs.New().RemoveAll("x")

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithRemoveAll", func() {
			called := false
			f := testfs.New(testfs.WithRemoveAll(func(string) error {
				called = true
				return nil
			}))

			_ = f.RemoveAll("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("ReadDir", func() {
		It("returns ErrNotExist by default", func() {
			_, err := testfs.New().ReadDir("x")

			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("calls custom function when set via WithReadDir", func() {
			called := false
			f := testfs.New(testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
				called = true
				return nil, nil
			}))

			_, _ = f.ReadDir("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("CreateTemp", func() {
		It("returns ErrPermission by default", func() {
			_, err := testfs.New().CreateTemp("x", "y")

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithCreateTemp", func() {
			called := false
			f := testfs.New(testfs.WithCreateTemp(func(string, string) (ihfs.File, error) {
				called = true
				return nil, nil
			}))

			_, _ = f.CreateTemp("x", "y")

			Expect(called).To(BeTrue())
		})
	})

	Describe("Glob", func() {
		It("returns ErrPermission by default", func() {
			_, err := testfs.New().Glob("x")

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithGlob", func() {
			called := false
			f := testfs.New(testfs.WithGlob(func(string) ([]string, error) {
				called = true
				return nil, nil
			}))

			_, _ = f.Glob("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("Lstat", func() {
		It("returns ErrNotExist by default", func() {
			_, err := testfs.New().Lstat("x")

			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("calls custom function when set via WithLstat", func() {
			called := false
			f := testfs.New(testfs.WithLstat(func(string) (ihfs.FileInfo, error) {
				called = true
				return nil, nil
			}))

			_, _ = f.Lstat("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("OpenFile", func() {
		It("returns ErrPermission by default", func() {
			_, err := testfs.New().OpenFile("x", 0, 0)

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithOpenFile", func() {
			called := false
			f := testfs.New(testfs.WithOpenFile(func(string, int, ihfs.FileMode) (ihfs.File, error) {
				called = true
				return nil, nil
			}))

			_, _ = f.OpenFile("x", 0, 0)

			Expect(called).To(BeTrue())
		})
	})

	Describe("ReadDirNames", func() {
		It("returns ErrNotExist by default", func() {
			_, err := testfs.New().ReadDirNames("x")

			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("calls custom function when set via WithReadDirNames", func() {
			called := false
			f := testfs.New(testfs.WithReadDirNames(func(string) ([]string, error) {
				called = true
				return nil, nil
			}))

			_, _ = f.ReadDirNames("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("ReadFile", func() {
		It("returns ErrPermission by default", func() {
			_, err := testfs.New().ReadFile("x")

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithReadFile", func() {
			called := false
			f := testfs.New(testfs.WithReadFile(func(string) ([]byte, error) {
				called = true
				return nil, nil
			}))

			_, _ = f.ReadFile("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("ReadLink", func() {
		It("returns ErrInvalid by default", func() {
			_, err := testfs.New().ReadLink("x")

			Expect(err).To(MatchError(fs.ErrInvalid))
		})

		It("calls custom function when set via WithReadLink", func() {
			called := false
			f := testfs.New(testfs.WithReadLink(func(string) (string, error) {
				called = true
				return "", nil
			}))

			_, _ = f.ReadLink("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("Rename", func() {
		It("returns ErrPermission by default", func() {
			err := testfs.New().Rename("x", "y")

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithRename", func() {
			called := false
			f := testfs.New(testfs.WithRename(func(string, string) error {
				called = true
				return nil
			}))

			_ = f.Rename("x", "y")

			Expect(called).To(BeTrue())
		})
	})

	Describe("Sub", func() {
		It("returns ErrNotExist by default", func() {
			_, err := testfs.New().Sub("x")

			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("calls custom function when set via WithSub", func() {
			called := false
			f := testfs.New(testfs.WithSub(func(string) (ihfs.FS, error) {
				called = true
				return nil, nil
			}))

			_, _ = f.Sub("x")

			Expect(called).To(BeTrue())
		})
	})

	Describe("Symlink", func() {
		It("returns ErrPermission by default", func() {
			err := testfs.New().Symlink("x", "y")

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithSymlink", func() {
			called := false
			f := testfs.New(testfs.WithSymlink(func(string, string) error {
				called = true
				return nil
			}))

			_ = f.Symlink("x", "y")

			Expect(called).To(BeTrue())
		})
	})

	Describe("TempFile", func() {
		It("returns ErrPermission by default", func() {
			_, err := testfs.New().TempFile("x", "y")

			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("calls custom function when set via WithTempFile", func() {
			called := false
			f := testfs.New(testfs.WithTempFile(func(string, string) (string, error) {
				called = true
				return "", nil
			}))

			_, _ = f.TempFile("x", "y")

			Expect(called).To(BeTrue())
		})
	})
})

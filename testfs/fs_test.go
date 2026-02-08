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

var _ = Describe("Fs", func() {
	var f *testfs.Fs

	BeforeEach(func() {
		f = testfs.New()
	})

	Describe("NewFs", func() {
		It("should create a new Fs", func() {
			Expect(f).NotTo(BeNil())
		})
	})

	Describe("Named", func() {
		It("should set the name", func() {
			f.Named("test")
			Expect(f.Name()).To(Equal("test"))
		})

		It("should return the Fs for chaining", func() {
			result := f.Named("test")
			Expect(result).To(Equal(f))
		})
	})

	Describe("Name", func() {
		It("should return the default name", func() {
			Expect(f.Name()).To(Equal("testfs"))
		})
	})

	Describe("Open", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.Open("test.txt")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			file := &testfs.File{}
			f.WithOpen(func(path string) (ihfs.File, error) {
				Expect(path).To(Equal("test.txt"))
				return file, nil
			})

			result, err := f.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(file))
		})

		It("should dequeue mocks in FIFO order", func() {
			file1 := &testfs.File{}
			file2 := &testfs.File{}
			f.WithOpen(
				func(path string) (ihfs.File, error) { return file1, nil },
				func(path string) (ihfs.File, error) { return file2, nil },
			)

			result1, err1 := f.Open("test1.txt")
			Expect(err1).NotTo(HaveOccurred())
			Expect(result1).To(Equal(file1))

			result2, err2 := f.Open("test2.txt")
			Expect(err2).NotTo(HaveOccurred())
			Expect(result2).To(Equal(file2))

			_, err3 := f.Open("test3.txt")
			Expect(err3).To(MatchError(testfs.ErrNoMocks))
		})
	})

	Describe("WithOpen", func() {
		It("should append mocks", func() {
			f.WithOpen(func(path string) (ihfs.File, error) { return nil, nil })
			f.WithOpen(func(path string) (ihfs.File, error) { return nil, nil })

			f.Open("test1.txt")
			_, err := f.Open("test2.txt")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the Fs for chaining", func() {
			result := f.WithOpen(func(path string) (ihfs.File, error) { return nil, nil })
			Expect(result).To(Equal(f))
		})
	})

	Describe("SetOpen", func() {
		It("should replace mocks", func() {
			f.WithOpen(func(path string) (ihfs.File, error) { return nil, errors.New("old") })
			f.SetOpen(func(path string) (ihfs.File, error) { return nil, errors.New("new") })

			_, err := f.Open("test.txt")
			Expect(err).To(MatchError("new"))
		})

		It("should return the Fs for chaining", func() {
			result := f.SetOpen(func(path string) (ihfs.File, error) { return nil, nil })
			Expect(result).To(Equal(f))
		})
	})

	Describe("Stat", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.Stat("test.txt")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			info := testfs.NewFileInfo("test.txt")
			f.WithStat(func(path string) (ihfs.FileInfo, error) {
				Expect(path).To(Equal("test.txt"))
				return info, nil
			})

			result, err := f.Stat("test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(info))
		})

		It("should dequeue mocks in FIFO order", func() {
			info1 := testfs.NewFileInfo("file1.txt")
			info2 := testfs.NewFileInfo("file2.txt")
			f.WithStat(
				func(path string) (ihfs.FileInfo, error) { return info1, nil },
				func(path string) (ihfs.FileInfo, error) { return info2, nil },
			)

			result1, _ := f.Stat("test1.txt")
			Expect(result1).To(Equal(info1))

			result2, _ := f.Stat("test2.txt")
			Expect(result2).To(Equal(info2))

			_, err := f.Stat("test3.txt")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})
	})

	Describe("WithStat", func() {
		It("should append mocks", func() {
			f.WithStat(func(path string) (ihfs.FileInfo, error) { return nil, nil })
			f.WithStat(func(path string) (ihfs.FileInfo, error) { return nil, nil })

			f.Stat("test1.txt")
			_, err := f.Stat("test2.txt")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the Fs for chaining", func() {
			result := f.WithStat(func(path string) (ihfs.FileInfo, error) { return nil, nil })
			Expect(result).To(Equal(f))
		})
	})

	Describe("SetStat", func() {
		It("should replace mocks", func() {
			f.WithStat(func(path string) (ihfs.FileInfo, error) { return nil, errors.New("old") })
			f.SetStat(func(path string) (ihfs.FileInfo, error) { return nil, errors.New("new") })

			_, err := f.Stat("test.txt")
			Expect(err).To(MatchError("new"))
		})

		It("should return the Fs for chaining", func() {
			result := f.SetStat(func(path string) (ihfs.FileInfo, error) { return nil, nil })
			Expect(result).To(Equal(f))
		})
	})

	Describe("Chmod", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			err := f.Chmod("test.txt", 0644)
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			f.WithChmod(func(name string, mode ihfs.FileMode) error {
				Expect(name).To(Equal("test.txt"))
				Expect(mode).To(Equal(ihfs.FileMode(0644)))
				return nil
			})

			err := f.Chmod("test.txt", 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should dequeue mocks in FIFO order", func() {
			f.WithChmod(
				func(name string, mode ihfs.FileMode) error { return nil },
				func(name string, mode ihfs.FileMode) error { return errors.New("second") },
			)

			err1 := f.Chmod("test.txt", 0644)
			Expect(err1).NotTo(HaveOccurred())

			err2 := f.Chmod("test.txt", 0755)
			Expect(err2).To(MatchError("second"))

			err3 := f.Chmod("test.txt", 0644)
			Expect(err3).To(MatchError(testfs.ErrNoMocks))
		})
	})

	Describe("WithChmod", func() {
		It("should return the Fs for chaining", func() {
			result := f.WithChmod(func(name string, mode ihfs.FileMode) error { return nil })
			Expect(result).To(Equal(f))
		})
	})

	Describe("SetChmod", func() {
		It("should replace mocks", func() {
			f.WithChmod(func(name string, mode ihfs.FileMode) error { return errors.New("old") })
			f.SetChmod(func(name string, mode ihfs.FileMode) error { return errors.New("new") })

			err := f.Chmod("test.txt", 0644)
			Expect(err).To(MatchError("new"))
		})

		It("should return the Fs for chaining", func() {
			result := f.SetChmod(func(name string, mode ihfs.FileMode) error { return nil })
			Expect(result).To(Equal(f))
		})
	})

	Describe("Chown", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			err := f.Chown("test.txt", 1000, 1000)
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			f.WithChown(func(name string, uid, gid int) error {
				Expect(name).To(Equal("test.txt"))
				Expect(uid).To(Equal(1000))
				Expect(gid).To(Equal(1000))
				return nil
			})

			err := f.Chown("test.txt", 1000, 1000)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("SetChown", func() {
		It("should replace mocks", func() {
			f.WithChown(func(name string, uid, gid int) error { return errors.New("old") })
			f.SetChown(func(name string, uid, gid int) error { return errors.New("new") })

			err := f.Chown("test.txt", 1000, 1000)
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("Chtimes", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			err := f.Chtimes("test.txt", time.Now(), time.Now())
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			atime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			mtime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
			f.WithChtimes(func(name string, at, mt time.Time) error {
				Expect(name).To(Equal("test.txt"))
				Expect(at).To(Equal(atime))
				Expect(mt).To(Equal(mtime))
				return nil
			})

			err := f.Chtimes("test.txt", atime, mtime)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("SetChtimes", func() {
		It("should replace mocks", func() {
			f.WithChtimes(func(name string, at, mt time.Time) error { return errors.New("old") })
			f.SetChtimes(func(name string, at, mt time.Time) error { return errors.New("new") })

			err := f.Chtimes("test.txt", time.Now(), time.Now())
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("Copy", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			err := f.Copy("dest", testfs.New())
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			srcFs := testfs.New()
			called := false
			f.WithCopy(func(dir string, fsys ihfs.FS) error {
				Expect(dir).To(Equal("dest"))
				called = true
				return nil
			})

			err := f.Copy("dest", srcFs)
			Expect(err).NotTo(HaveOccurred())
			Expect(called).To(BeTrue())
		})
	})

	Describe("Create", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.Create("test.txt")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			file := &testfs.File{}
			f.WithCreate(func(name string) (ihfs.File, error) {
				Expect(name).To(Equal("test.txt"))
				return file, nil
			})

			result, err := f.Create("test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(file))
		})
	})

	Describe("CreateTemp", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.CreateTemp("dir", "pattern")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			file := &testfs.File{}
			f.WithCreateTemp(func(dir, pattern string) (ihfs.File, error) {
				Expect(dir).To(Equal("dir"))
				Expect(pattern).To(Equal("pattern"))
				return file, nil
			})

			result, err := f.CreateTemp("dir", "pattern")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(file))
		})
	})

	Describe("Glob", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.Glob("*.txt")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			matches := []string{"a.txt", "b.txt"}
			f.WithGlob(func(pattern string) ([]string, error) {
				Expect(pattern).To(Equal("*.txt"))
				return matches, nil
			})

			result, err := f.Glob("*.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(matches))
		})
	})

	Describe("Lstat", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.Lstat("test.txt")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			info := testfs.NewFileInfo("test.txt")
			f.WithLstat(func(name string) (ihfs.FileInfo, error) {
				Expect(name).To(Equal("test.txt"))
				return info, nil
			})

			result, err := f.Lstat("test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(info))
		})
	})

	Describe("Mkdir", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			err := f.Mkdir("dir", 0755)
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			f.WithMkdir(func(name string, mode ihfs.FileMode) error {
				Expect(name).To(Equal("dir"))
				Expect(mode).To(Equal(ihfs.FileMode(0755)))
				return nil
			})

			err := f.Mkdir("dir", 0755)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("MkdirAll", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			err := f.MkdirAll("path/to/dir", 0755)
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			f.WithMkdirAll(func(name string, mode ihfs.FileMode) error {
				Expect(name).To(Equal("path/to/dir"))
				Expect(mode).To(Equal(ihfs.FileMode(0755)))
				return nil
			})

			err := f.MkdirAll("path/to/dir", 0755)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("MkdirTemp", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.MkdirTemp("dir", "pattern")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			f.WithMkdirTemp(func(dir, pattern string) (string, error) {
				Expect(dir).To(Equal("dir"))
				Expect(pattern).To(Equal("pattern"))
				return "/tmp/test123", nil
			})

			result, err := f.MkdirTemp("dir", "pattern")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("/tmp/test123"))
		})
	})

	Describe("OpenFile", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.OpenFile("test.txt", 0, 0644)
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			file := &testfs.File{}
			f.WithOpenFile(func(name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
				Expect(name).To(Equal("test.txt"))
				Expect(flag).To(Equal(0))
				Expect(perm).To(Equal(ihfs.FileMode(0644)))
				return file, nil
			})

			result, err := f.OpenFile("test.txt", 0, 0644)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(file))
		})
	})

	Describe("ReadDir", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.ReadDir("dir")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			entries := []ihfs.DirEntry{
				fs.FileInfoToDirEntry(testfs.NewFileInfo("a.txt")),
				fs.FileInfoToDirEntry(testfs.NewFileInfo("b.txt")),
			}
			f.WithReadDir(func(name string) ([]ihfs.DirEntry, error) {
				Expect(name).To(Equal("dir"))
				return entries, nil
			})

			result, err := f.ReadDir("dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(entries))
		})
	})

	Describe("ReadDirNames", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.ReadDirNames("dir")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			names := []string{"a.txt", "b.txt"}
			f.WithReadDirNames(func(name string) ([]string, error) {
				Expect(name).To(Equal("dir"))
				return names, nil
			})

			result, err := f.ReadDirNames("dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(names))
		})
	})

	Describe("ReadFile", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.ReadFile("test.txt")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			data := []byte("test content")
			f.WithReadFile(func(name string) ([]byte, error) {
				Expect(name).To(Equal("test.txt"))
				return data, nil
			})

			result, err := f.ReadFile("test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(data))
		})
	})

	Describe("ReadLink", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.ReadLink("link")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			f.WithReadLink(func(name string) (string, error) {
				Expect(name).To(Equal("link"))
				return "target", nil
			})

			result, err := f.ReadLink("link")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("target"))
		})
	})

	Describe("Remove", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			err := f.Remove("test.txt")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			f.WithRemove(func(name string) error {
				Expect(name).To(Equal("test.txt"))
				return nil
			})

			err := f.Remove("test.txt")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("RemoveAll", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			err := f.RemoveAll("dir")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			f.WithRemoveAll(func(name string) error {
				Expect(name).To(Equal("dir"))
				return nil
			})

			err := f.RemoveAll("dir")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Rename", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			err := f.Rename("old", "new")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			f.WithRename(func(oldpath, newpath string) error {
				Expect(oldpath).To(Equal("old"))
				Expect(newpath).To(Equal("new"))
				return nil
			})

			err := f.Rename("old", "new")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Sub", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.Sub("dir")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			subFs := testfs.New()
			called := false
			f.WithSub(func(dir string) (ihfs.FS, error) {
				Expect(dir).To(Equal("dir"))
				called = true
				return subFs, nil
			})

			result, err := f.Sub("dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(called).To(BeTrue())
		})
	})

	Describe("Symlink", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			err := f.Symlink("target", "link")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			f.WithSymlink(func(oldname, newname string) error {
				Expect(oldname).To(Equal("target"))
				Expect(newname).To(Equal("link"))
				return nil
			})

			err := f.Symlink("target", "link")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("TempFile", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			_, err := f.TempFile("dir", "pattern")
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			f.WithTempFile(func(dir, pattern string) (string, error) {
				Expect(dir).To(Equal("dir"))
				Expect(pattern).To(Equal("pattern"))
				return "/tmp/test123", nil
			})

			result, err := f.TempFile("dir", "pattern")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("/tmp/test123"))
		})
	})

	Describe("WriteFile", func() {
		It("should return ErrNoMocks when no mocks are set", func() {
			err := f.WriteFile("test.txt", []byte("data"), 0644)
			Expect(err).To(MatchError(testfs.ErrNoMocks))
		})

		It("should execute the first mock", func() {
			data := []byte("test data")
			f.WithWriteFile(func(name string, d []byte, perm ihfs.FileMode) error {
				Expect(name).To(Equal("test.txt"))
				Expect(d).To(Equal(data))
				Expect(perm).To(Equal(ihfs.FileMode(0644)))
				return nil
			})

			err := f.WriteFile("test.txt", data, 0644)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("SetCopy", func() {
		It("should replace mocks", func() {
			f.WithCopy(func(dir string, fsys ihfs.FS) error { return errors.New("old") })
			f.SetCopy(func(dir string, fsys ihfs.FS) error { return errors.New("new") })

			err := f.Copy("dest", testfs.New())
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetCreate", func() {
		It("should replace mocks", func() {
			f.WithCreate(func(name string) (ihfs.File, error) { return nil, errors.New("old") })
			f.SetCreate(func(name string) (ihfs.File, error) { return nil, errors.New("new") })

			_, err := f.Create("test.txt")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetCreateTemp", func() {
		It("should replace mocks", func() {
			f.WithCreateTemp(func(dir, pattern string) (ihfs.File, error) { return nil, errors.New("old") })
			f.SetCreateTemp(func(dir, pattern string) (ihfs.File, error) { return nil, errors.New("new") })

			_, err := f.CreateTemp("dir", "pattern")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetGlob", func() {
		It("should replace mocks", func() {
			f.WithGlob(func(pattern string) ([]string, error) { return nil, errors.New("old") })
			f.SetGlob(func(pattern string) ([]string, error) { return nil, errors.New("new") })

			_, err := f.Glob("*.txt")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetLstat", func() {
		It("should replace mocks", func() {
			f.WithLstat(func(name string) (ihfs.FileInfo, error) { return nil, errors.New("old") })
			f.SetLstat(func(name string) (ihfs.FileInfo, error) { return nil, errors.New("new") })

			_, err := f.Lstat("test.txt")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetMkdir", func() {
		It("should replace mocks", func() {
			f.WithMkdir(func(name string, mode ihfs.FileMode) error { return errors.New("old") })
			f.SetMkdir(func(name string, mode ihfs.FileMode) error { return errors.New("new") })

			err := f.Mkdir("dir", 0755)
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetMkdirAll", func() {
		It("should replace mocks", func() {
			f.WithMkdirAll(func(name string, mode ihfs.FileMode) error { return errors.New("old") })
			f.SetMkdirAll(func(name string, mode ihfs.FileMode) error { return errors.New("new") })

			err := f.MkdirAll("path/to/dir", 0755)
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetMkdirTemp", func() {
		It("should replace mocks", func() {
			f.WithMkdirTemp(func(dir, pattern string) (string, error) { return "", errors.New("old") })
			f.SetMkdirTemp(func(dir, pattern string) (string, error) { return "", errors.New("new") })

			_, err := f.MkdirTemp("dir", "pattern")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetOpenFile", func() {
		It("should replace mocks", func() {
			f.WithOpenFile(func(name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
				return nil, errors.New("old")
			})
			f.SetOpenFile(func(name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
				return nil, errors.New("new")
			})

			_, err := f.OpenFile("test.txt", 0, 0644)
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetReadDir", func() {
		It("should replace mocks", func() {
			f.WithReadDir(func(name string) ([]ihfs.DirEntry, error) { return nil, errors.New("old") })
			f.SetReadDir(func(name string) ([]ihfs.DirEntry, error) { return nil, errors.New("new") })

			_, err := f.ReadDir("dir")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetReadDirNames", func() {
		It("should replace mocks", func() {
			f.WithReadDirNames(func(name string) ([]string, error) { return nil, errors.New("old") })
			f.SetReadDirNames(func(name string) ([]string, error) { return nil, errors.New("new") })

			_, err := f.ReadDirNames("dir")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetReadFile", func() {
		It("should replace mocks", func() {
			f.WithReadFile(func(name string) ([]byte, error) { return nil, errors.New("old") })
			f.SetReadFile(func(name string) ([]byte, error) { return nil, errors.New("new") })

			_, err := f.ReadFile("test.txt")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetReadLink", func() {
		It("should replace mocks", func() {
			f.WithReadLink(func(name string) (string, error) { return "", errors.New("old") })
			f.SetReadLink(func(name string) (string, error) { return "", errors.New("new") })

			_, err := f.ReadLink("link")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetRemove", func() {
		It("should replace mocks", func() {
			f.WithRemove(func(name string) error { return errors.New("old") })
			f.SetRemove(func(name string) error { return errors.New("new") })

			err := f.Remove("test.txt")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetRemoveAll", func() {
		It("should replace mocks", func() {
			f.WithRemoveAll(func(name string) error { return errors.New("old") })
			f.SetRemoveAll(func(name string) error { return errors.New("new") })

			err := f.RemoveAll("dir")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetRename", func() {
		It("should replace mocks", func() {
			f.WithRename(func(oldpath, newpath string) error { return errors.New("old") })
			f.SetRename(func(oldpath, newpath string) error { return errors.New("new") })

			err := f.Rename("old", "new")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetSub", func() {
		It("should replace mocks", func() {
			f.WithSub(func(dir string) (ihfs.FS, error) { return nil, errors.New("old") })
			f.SetSub(func(dir string) (ihfs.FS, error) { return nil, errors.New("new") })

			_, err := f.Sub("dir")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetSymlink", func() {
		It("should replace mocks", func() {
			f.WithSymlink(func(oldname, newname string) error { return errors.New("old") })
			f.SetSymlink(func(oldname, newname string) error { return errors.New("new") })

			err := f.Symlink("target", "link")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetTempFile", func() {
		It("should replace mocks", func() {
			f.WithTempFile(func(dir, pattern string) (string, error) { return "", errors.New("old") })
			f.SetTempFile(func(dir, pattern string) (string, error) { return "", errors.New("new") })

			_, err := f.TempFile("dir", "pattern")
			Expect(err).To(MatchError("new"))
		})
	})

	Describe("SetWriteFile", func() {
		It("should replace mocks", func() {
			f.WithWriteFile(func(name string, data []byte, perm ihfs.FileMode) error {
				return errors.New("old")
			})
			f.SetWriteFile(func(name string, data []byte, perm ihfs.FileMode) error {
				return errors.New("new")
			})

			err := f.WriteFile("test.txt", []byte("data"), 0644)
			Expect(err).To(MatchError("new"))
		})
	})
})

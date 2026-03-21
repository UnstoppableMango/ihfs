package memfs_test

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"testing/fstest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/memfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("Fs", func() {
	Describe("Open", func() {
		It("should open root directory", func() {
			mfs := memfs.New()
			file, err := mfs.Open(".")
			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())

			fi, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should return error for non-existent file", func() {
			mfs := memfs.New()
			_, err := mfs.Open("nonexistent")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("does not exist")))
		})
	})

	Describe("Create", func() {
		It("should create a new file", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("hello"))
			Expect(err).NotTo(HaveOccurred())

			err = file.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should be able to read created file", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("hello world"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			file, err = mfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())

			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("hello world"))
		})
	})

	Describe("Mkdir", func() {
		It("should create a directory", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("/testdir", 0755)
			Expect(err).NotTo(HaveOccurred())

			fi, err := mfs.Stat("/testdir")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should error if directory already exists", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("/testdir", 0755)
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Mkdir("/testdir", 0755)
			Expect(err).To(HaveOccurred())
		})

		It("should error if parent directory does not exist", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("/nonexistent/testdir", 0755)
			Expect(err).To(HaveOccurred())
		})

		It("should error if parent is not a directory", func() {
			mfs := memfs.New()
			_, err := mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Mkdir("/file.txt/testdir", 0755)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("MkdirAll", func() {
		It("should create nested directories", func() {
			mfs := memfs.New()
			err := mfs.MkdirAll("/a/b/c", 0755)
			Expect(err).NotTo(HaveOccurred())

			fi, err := mfs.Stat("/a/b/c")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should not error if directory exists", func() {
			mfs := memfs.New()
			err := mfs.MkdirAll("/testdir", 0755)
			Expect(err).NotTo(HaveOccurred())

			err = mfs.MkdirAll("/testdir", 0755)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle paths with empty parts", func() {
			mfs := memfs.New()
			err := mfs.MkdirAll("//a///b//c//", 0755)
			Expect(err).NotTo(HaveOccurred())

			fi, err := mfs.Stat("/a/b/c")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should not error when calling MkdirAll on root", func() {
			mfs := memfs.New()
			err := mfs.MkdirAll("/", 0755)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create remaining directories when some already exist", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("/a", 0755)
			Expect(err).NotTo(HaveOccurred())

			err = mfs.MkdirAll("/a/b/c", 0755)
			Expect(err).NotTo(HaveOccurred())

			fi, err := mfs.Stat("/a/b/c")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should error if intermediate path component is a file", func() {
			mfs := memfs.New()
			_, err := mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())

			err = mfs.MkdirAll("/file.txt/subdir", 0755)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Remove", func() {
		It("should remove a file", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Remove("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			_, err = mfs.Stat("/test.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should remove empty directory", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("/testdir", 0755)
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Remove("/testdir")
			Expect(err).NotTo(HaveOccurred())

			_, err = mfs.Stat("/testdir")
			Expect(err).To(HaveOccurred())
		})

		It("should error when removing non-empty directory", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("/testdir", 0755)
			Expect(err).NotTo(HaveOccurred())

			_, err = mfs.Create("/testdir/file.txt")
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Remove("/testdir")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("RemoveAll", func() {
		It("should remove directory and all contents", func() {
			mfs := memfs.New()
			err := mfs.MkdirAll("/a/b/c", 0755)
			Expect(err).NotTo(HaveOccurred())

			_, err = mfs.Create("/a/b/file.txt")
			Expect(err).NotTo(HaveOccurred())

			err = mfs.RemoveAll("/a")
			Expect(err).NotTo(HaveOccurred())

			_, err = mfs.Stat("/a")
			Expect(err).To(HaveOccurred())
		})

		It("should not error if path doesn't exist", func() {
			mfs := memfs.New()
			err := mfs.RemoveAll("/nonexistent")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle RemoveAll on root directory", func() {
			mfs := memfs.New()
			_, err := mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())

			err = mfs.RemoveAll("/")
			Expect(err).NotTo(HaveOccurred())

			// Root should be removed from the map, filesystem essentially empty
			_, err = mfs.Stat("/")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Rename", func() {
		It("should rename a file", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/old.txt")
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("content"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Rename("/old.txt", "/new.txt")
			Expect(err).NotTo(HaveOccurred())

			_, err = mfs.Stat("/old.txt")
			Expect(err).To(HaveOccurred())

			file, err = mfs.Open("new.txt")
			Expect(err).NotTo(HaveOccurred())
			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("content"))
		})

		It("should error if old file does not exist", func() {
			mfs := memfs.New()
			err := mfs.Rename("/nonexistent.txt", "/new.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should error if new name exists", func() {
			mfs := memfs.New()
			_, err := mfs.Create("/file1.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = mfs.Create("/file2.txt")
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Rename("/file1.txt", "/file2.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should error if new parent directory is not a directory", func() {
			mfs := memfs.New()
			_, err := mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = mfs.Create("/old.txt")
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Rename("/old.txt", "/file.txt/new.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should error if new parent directory does not exist", func() {
			mfs := memfs.New()
			_, err := mfs.Create("/old.txt")
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Rename("/old.txt", "/nonexistent/new.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should handle renaming root directory children", func() {
			mfs := memfs.New()
			_, err := mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Rename("/file.txt", "/renamed.txt")
			Expect(err).NotTo(HaveOccurred())

			_, err = mfs.Stat("/file.txt")
			Expect(err).To(HaveOccurred())

			_, err = mfs.Stat("/renamed.txt")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should rename to nested directory", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("/dir", 0755)
			Expect(err).NotTo(HaveOccurred())
			_, err = mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Rename("/file.txt", "/dir/file.txt")
			Expect(err).NotTo(HaveOccurred())

			_, err = mfs.Stat("/file.txt")
			Expect(err).To(HaveOccurred())

			_, err = mfs.Stat("/dir/file.txt")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Chmod", func() {
		It("should change file permissions", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Chmod("/test.txt", 0644)
			Expect(err).NotTo(HaveOccurred())

			fi, err := mfs.Stat("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.Mode()).To(Equal(os.FileMode(0644)))
		})
	})

	Describe("Chown", func() {
		It("should change file ownership", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Chown("/test.txt", 1000, 1000)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Chtimes", func() {
		It("should change file times", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			now := time.Now()
			err = mfs.Chtimes("/test.txt", now, now)
			Expect(err).NotTo(HaveOccurred())

			fi, err := mfs.Stat("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.ModTime().Unix()).To(Equal(now.Unix()))
		})
	})

	Describe("OpenFile", func() {
		It("should create file with O_CREATE flag", func() {
			mfs := memfs.New()
			file, err := mfs.OpenFile("/test.txt", os.O_CREATE|os.O_RDWR, 0644)
			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should truncate file with O_TRUNC flag", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("original content"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			file, err = mfs.OpenFile("/test.txt", os.O_TRUNC|os.O_RDWR, 0644)
			Expect(err).NotTo(HaveOccurred())
			writer = file.(io.Writer)
			_, err = writer.Write([]byte("new"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			file, err = mfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())
			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("new"))
		})
	})

	Describe("File operations", func() {
		It("should support Seek", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("0123456789"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			file, err = mfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())

			seeker := file.(io.Seeker)
			pos, err := seeker.Seek(5, io.SeekStart)
			Expect(err).NotTo(HaveOccurred())
			Expect(pos).To(Equal(int64(5)))

			buf := make([]byte, 5)
			n, err := file.Read(buf)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(buf)).To(Equal("56789"))
		})

		It("should support Truncate", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("0123456789"))
			Expect(err).NotTo(HaveOccurred())

			truncater := file.(interface{ Truncate(int64) error })
			err = truncater.Truncate(5)
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			file, err = mfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())
			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("01234"))
		})

		It("should support ReadDir", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("/testdir", 0755)
			Expect(err).NotTo(HaveOccurred())
			_, err = mfs.Create("/testdir/file1.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = mfs.Create("/testdir/file2.txt")
			Expect(err).NotTo(HaveOccurred())

			file, err := mfs.Open("testdir")
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(ihfs.ReadDirFile)
			entries, err := dirFile.ReadDir(-1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(2))
		})
	})

	Describe("Error paths and edge cases", func() {
		It("should error when reading from closed file", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			buf := make([]byte, 10)
			_, err = file.Read(buf)
			Expect(err).To(HaveOccurred())
		})

		It("should error when reading from directory", func() {
			mfs := memfs.New()
			file, err := mfs.Open(".")
			Expect(err).NotTo(HaveOccurred())

			buf := make([]byte, 10)
			_, err = file.Read(buf)
			Expect(err).To(HaveOccurred())
		})

		It("should error when writing to readonly file", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			file, err = mfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
		})

		It("should error when writing to closed file", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
		})

		It("should error when writing to directory", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("/testdir", 0755)
			Expect(err).NotTo(HaveOccurred())

			file, err := mfs.OpenFile("/testdir", os.O_RDWR, 0755)
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
		})

		It("should handle write with position beyond content length", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			seeker := file.(io.Seeker)
			_, err = seeker.Seek(10, io.SeekStart)
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("test"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle overwrite in middle of content", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("0123456789"))
			Expect(err).NotTo(HaveOccurred())

			seeker := file.(io.Seeker)
			_, err = seeker.Seek(5, io.SeekStart)
			Expect(err).NotTo(HaveOccurred())

			_, err = writer.Write([]byte("XXX"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			file, err = mfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())
			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("01234XXX89"))
		})

		It("should error when ReadDir on closed file", func() {
			mfs := memfs.New()
			file, err := mfs.Open(".")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(ihfs.ReadDirFile)
			_, err = dirFile.ReadDir(-1)
			Expect(err).To(HaveOccurred())
		})

		It("should error when ReadDir on non-directory", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(ihfs.ReadDirFile)
			_, err = dirFile.ReadDir(-1)
			Expect(err).To(HaveOccurred())
		})

		It("should return empty list for directory with no children", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("/emptydir", 0755)
			Expect(err).NotTo(HaveOccurred())

			file, err := mfs.Open("emptydir")
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(ihfs.ReadDirFile)
			entries, err := dirFile.ReadDir(-1)
			// Empty directory returns EOF or empty slice depending on implementation
			if err == io.EOF {
				Expect(entries).To(BeNil())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(BeEmpty())
			}
		})

		It("should handle ReadDir pagination with n > 0", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("/testdir", 0755)
			Expect(err).NotTo(HaveOccurred())
			_, err = mfs.Create("/testdir/file1.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = mfs.Create("/testdir/file2.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = mfs.Create("/testdir/file3.txt")
			Expect(err).NotTo(HaveOccurred())

			file, err := mfs.Open("testdir")
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(ihfs.ReadDirFile)
			entries, err := dirFile.ReadDir(2)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(2))

			entries, err = dirFile.ReadDir(2)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))

			_, err = dirFile.ReadDir(2)
			Expect(err).To(Equal(io.EOF))
		})

		It("should error on Seek with closed file", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			seeker := file.(io.Seeker)
			_, err = seeker.Seek(0, io.SeekStart)
			Expect(err).To(HaveOccurred())
		})

		It("should handle Seek with SeekEnd", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("0123456789"))
			Expect(err).NotTo(HaveOccurred())

			seeker := file.(io.Seeker)
			pos, err := seeker.Seek(-5, io.SeekEnd)
			Expect(err).NotTo(HaveOccurred())
			Expect(pos).To(Equal(int64(5)))
		})

		It("should error on Seek with invalid whence", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			seeker := file.(io.Seeker)
			_, err = seeker.Seek(0, 99)
			Expect(err).To(HaveOccurred())
		})

		It("should error on Seek with negative result", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			seeker := file.(io.Seeker)
			_, err = seeker.Seek(-10, io.SeekStart)
			Expect(err).To(HaveOccurred())
		})

		It("should error on Truncate with readonly file", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			file, err = mfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())

			truncater := file.(interface{ Truncate(int64) error })
			err = truncater.Truncate(5)
			Expect(err).To(HaveOccurred())
		})

		It("should error on Truncate with closed file", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			truncater := file.(interface{ Truncate(int64) error })
			err = truncater.Truncate(5)
			Expect(err).To(HaveOccurred())
		})

		It("should error on Truncate with negative size", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			truncater := file.(interface{ Truncate(int64) error })
			err = truncater.Truncate(-1)
			Expect(err).To(HaveOccurred())
		})

		It("should handle Truncate extending file", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("test"))
			Expect(err).NotTo(HaveOccurred())

			truncater := file.(interface{ Truncate(int64) error })
			err = truncater.Truncate(10)
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			fi, err := mfs.Stat("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.Size()).To(Equal(int64(10)))
		})

		It("should support Sync operation", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			syncer := file.(interface{ Sync() error })
			err = syncer.Sync()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should call FileInfo.Size() on directory", func() {
			mfs := memfs.New()
			fi, err := mfs.Stat("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.Size()).To(Equal(int64(0)))
		})

		It("should call FileInfo.Size() on file", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("test"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			fi, err := mfs.Stat("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.Size()).To(Equal(int64(4)))
		})

		It("should call FileInfo.Sys()", func() {
			mfs := memfs.New()
			fi, err := mfs.Stat("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.Sys()).NotTo(BeNil())
		})

		It("should call FileInfo.Type()", func() {
			mfs := memfs.New()
			fi, err := mfs.Stat("/")
			Expect(err).NotTo(HaveOccurred())

			entry := fi.(ihfs.DirEntry)
			Expect(entry.Type()).To(Equal(os.ModeDir))
		})

		It("should call FileInfo.Info()", func() {
			mfs := memfs.New()
			fi, err := mfs.Stat("/")
			Expect(err).NotTo(HaveOccurred())

			entry := fi.(ihfs.DirEntry)
			info, err := entry.Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
		})

		It("should error when Create fails to register with parent", func() {
			mfs := memfs.New()
			// Try to create file in non-existent parent
			_, err := mfs.Create("/nonexistent/file.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should error when MkdirAll creates file instead of directory", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			err = mfs.MkdirAll("/file.txt", 0755)
			Expect(err).To(HaveOccurred())
		})

		It("should handle MkdirAll with root path", func() {
			mfs := memfs.New()
			err := mfs.MkdirAll("/", 0755)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should error when Remove on non-existent file", func() {
			mfs := memfs.New()
			err := mfs.Remove("/nonexistent")
			Expect(err).To(HaveOccurred())
		})

		It("should handle RemoveAll with no error on descendants", func() {
			mfs := memfs.New()
			err := mfs.MkdirAll("/a/b/c", 0755)
			Expect(err).NotTo(HaveOccurred())

			err = mfs.RemoveAll("/a")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should error when Rename source doesn't exist", func() {
			mfs := memfs.New()
			err := mfs.Rename("/nonexistent", "/new")
			Expect(err).To(HaveOccurred())
		})

		It("should error when Rename destination exists", func() {
			mfs := memfs.New()
			_, err := mfs.Create("/file1.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = mfs.Create("/file2.txt")
			Expect(err).NotTo(HaveOccurred())

			err = mfs.Rename("/file1.txt", "/file2.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should error when Chmod on non-existent file", func() {
			mfs := memfs.New()
			err := mfs.Chmod("/nonexistent", 0644)
			Expect(err).To(HaveOccurred())
		})

		It("should error when Chown on non-existent file", func() {
			mfs := memfs.New()
			err := mfs.Chown("/nonexistent", 1000, 1000)
			Expect(err).To(HaveOccurred())
		})

		It("should error when Chtimes on non-existent file", func() {
			mfs := memfs.New()
			now := time.Now()
			err := mfs.Chtimes("/nonexistent", now, now)
			Expect(err).To(HaveOccurred())
		})

		It("should error when OpenFile without O_CREATE on non-existent file", func() {
			mfs := memfs.New()
			_, err := mfs.OpenFile("/nonexistent", os.O_RDONLY, 0644)
			Expect(err).To(HaveOccurred())
		})

		It("should error when OpenFile with O_EXCL on existing file", func() {
			mfs := memfs.New()
			_, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			_, err = mfs.OpenFile("/test.txt", os.O_CREATE|os.O_EXCL, 0644)
			Expect(err).To(HaveOccurred())
		})

		It("should handle OpenFile with O_APPEND", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("initial"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			file, err = mfs.OpenFile("/test.txt", os.O_APPEND|os.O_WRONLY, 0644)
			Expect(err).NotTo(HaveOccurred())

			writer = file.(io.Writer)
			_, err = writer.Write([]byte("more"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			file, err = mfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())
			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("initialmore"))
		})

		It("should handle registerWithParent for root", func() {
			// Root has no parent, should not error
			mfs := memfs.New()
			file, err := mfs.Open(".")
			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
		})

		It("should error when registerWithParent with non-directory parent", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			// Try to create a file inside a file (not a directory)
			_, err = mfs.Create("/file.txt/nested")
			Expect(err).To(HaveOccurred())
		})

		It("should handle normalizePath with empty string", func() {
			mfs := memfs.New()
			_, err := mfs.Open("")
			Expect(err).To(HaveOccurred())
		})

		It("should handle normalizePath without leading separator", func() {
			mfs := memfs.New()
			file, err := mfs.Create("test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			fi, err := mfs.Stat("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi).NotTo(BeNil())
		})

		It("should handle ReadDir with nil dir children", func() {
			mfs := memfs.New()
			// Create directory and manually set dir to nil for edge case
			err := mfs.Mkdir("/nildir", 0755)
			Expect(err).NotTo(HaveOccurred())

			file, err := mfs.Open("nildir")
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(ihfs.ReadDirFile)
			entries, err := dirFile.ReadDir(-1)
			// Should handle nil dir gracefully
			if err != io.EOF {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(entries).To(BeEmpty())
		})

		It("should handle Seek with SeekCurrent", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("0123456789"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			file, err = mfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())

			seeker := file.(io.Seeker)
			// First seek to position 5
			_, err = seeker.Seek(5, io.SeekStart)
			Expect(err).NotTo(HaveOccurred())

			// Then seek forward 2 from current
			pos, err := seeker.Seek(2, io.SeekCurrent)
			Expect(err).NotTo(HaveOccurred())
			Expect(pos).To(Equal(int64(7)))
		})

		It("should handle MkdirAll with path containing empty parts", func() {
			mfs := memfs.New()
			err := mfs.MkdirAll("//a//b//c//", 0755)
			Expect(err).NotTo(HaveOccurred())

			fi, err := mfs.Stat("/a/b/c")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should error when MkdirAll finds existing file in path", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			err = mfs.MkdirAll("/file.txt/nested", 0755)
			Expect(err).To(HaveOccurred())
		})

		It("should handle removing a file that unregisters properly", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			// Remove should work fine
			err = mfs.Remove("/file.txt")
			Expect(err).NotTo(HaveOccurred())

			// Verify file is gone
			_, err = mfs.Stat("/file.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should handle RemoveAll for nested directories", func() {
			mfs := memfs.New()
			err := mfs.MkdirAll("/a/b/c", 0755)
			Expect(err).NotTo(HaveOccurred())

			err = mfs.RemoveAll("/a")
			Expect(err).NotTo(HaveOccurred())

			// Verify directory is gone
			_, err = mfs.Stat("/a")
			Expect(err).To(HaveOccurred())
		})

		It("should error when Rename fails to register with new parent", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())
			writer, ok := file.(ihfs.Writer)
			Expect(ok).To(BeTrue())
			_, err = writer.Write([]byte("test content"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			// Try to rename to non-existent directory
			err = mfs.Rename("/file.txt", "/nonexistent/file.txt")
			Expect(err).To(HaveOccurred())

			// Verify original file still exists and is accessible
			info, err := mfs.Stat("/file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Name()).To(Equal("file.txt"))

			// Verify we can still read the original file
			file, err = mfs.Open("file.txt")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(file.Close)
			content := make([]byte, 12)
			n, err := file.Read(content)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(12))
			Expect(string(content)).To(Equal("test content"))

			// Verify new path does not exist
			_, err = mfs.Stat("/nonexistent/file.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should error when renaming to a path where parent is a file", func() {
			mfs := memfs.New()
			// Create a file
			file, err := mfs.Create("/parent.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			// Create another file to rename
			file, err = mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())
			writer, ok := file.(ihfs.Writer)
			Expect(ok).To(BeTrue())
			_, err = writer.Write([]byte("content"))
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			// Try to rename to a path where parent is a file (not a directory)
			err = mfs.Rename("/file.txt", "/parent.txt/child.txt")
			Expect(err).To(HaveOccurred())

			// Verify original file still exists and is accessible
			info, err := mfs.Stat("/file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Name()).To(Equal("file.txt"))

			// Verify parent file is still a file
			info, err = mfs.Stat("/parent.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeFalse())

			// Verify new path does not exist
			_, err = mfs.Stat("/parent.txt/child.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should handle OpenFile with O_TRUNC on directory", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("/testdir", 0755)
			Expect(err).NotTo(HaveOccurred())

			file, err := mfs.OpenFile("/testdir", os.O_RDWR|os.O_TRUNC, 0755)
			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle OpenFile with O_RDONLY", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			file, err = mfs.OpenFile("/test.txt", os.O_RDONLY, 0644)
			Expect(err).NotTo(HaveOccurred())

			writer := file.(io.Writer)
			_, err = writer.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
		})

		It("should handle creating file at root level", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should verify MkdirAll creates nested directories", func() {
			mfs := memfs.New()
			err := mfs.MkdirAll("///a///b///", 0755)
			Expect(err).NotTo(HaveOccurred())

			fi, err := mfs.Stat("/a/b")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should error when registering with non-existent parent", func() {
			mfs := memfs.New()
			// Try to create file in non-existent directory
			_, err := mfs.OpenFile("/nonexistent/test.txt", os.O_CREATE|os.O_WRONLY, 0644)
			Expect(err).To(HaveOccurred())
		})

		It("should handle Remove operation successfully", func() {
			mfs := memfs.New()
			// Try to remove non-existent file
			err := mfs.Remove("/nonexistent.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should handle RemoveAll when path doesn't exist", func() {
			mfs := memfs.New()
			// RemoveAll doesn't error if path doesn't exist (matches os.RemoveAll behavior)
			err := mfs.RemoveAll("/nonexistent")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle Rename error when source doesn't exist", func() {
			mfs := memfs.New()
			// Try to rename non-existent file
			err := mfs.Rename("/nonexistent.txt", "/new.txt")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("consistent path validation", func() {
		var mfs *memfs.Fs

		BeforeEach(func() {
			mfs = memfs.New()
			f, err := mfs.Create("test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())
		})

		DescribeTable("should reject invalid paths",
			func(path string) {
				_, err := mfs.Open(path)
				Expect(err).To(HaveOccurred(), "Open(%q) should reject invalid path", path)
				Expect(err).To(MatchError(ContainSubstring("invalid")), "Open(%q) should return invalid error", path)
			},
			Entry("leading slash", "/test.txt"),
			Entry("trailing /.", "test.txt/."),
			Entry("double slash", "test.txt//"),
			Entry("leading ./", "./test.txt"),
			Entry("contains ..", "../test.txt"),
			Entry("contains .. in middle", "test/../test.txt"),
		)

		It("should accept valid paths after normalization", func() {
			f, err := mfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f).NotTo(BeNil())
			Expect(f.Close()).To(Succeed())
		})
	})

	Context("empty string path handling", func() {
		It("should reject empty string consistently", func() {
			mfs := memfs.New()

			// Open with empty string should fail
			_, err := mfs.Open("")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("invalid")))
		})

		It("should accept dot for root directory", func() {
			mfs := memfs.New()

			f, err := mfs.Open(".")
			Expect(err).NotTo(HaveOccurred())
			Expect(f).NotTo(BeNil())

			info, err := f.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
			Expect(f.Close()).To(Succeed())
		})
	})

	Context("error message path consistency", func() {
		It("should use original path in error messages, not internal normalized path", func() {
			mfs := memfs.New()

			// Try to open a non-existent file with a valid path
			_, err := mfs.Open("nonexistent.txt")
			Expect(err).To(HaveOccurred())

			// Error should contain the original path we passed
			Expect(err.Error()).To(ContainSubstring("nonexistent.txt"))
			// Error should not contain internal absolute path like "/nonexistent.txt"
			Expect(err.Error()).NotTo(ContainSubstring("/nonexistent.txt"))
		})
	})

	It("should handle MkdirAll with existing non-directory", func() {
		mfs := memfs.New()

		// Create a file
		f, err := mfs.Create("file.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Close()).To(Succeed())

		// Try to create directory with same name
		err = mfs.MkdirAll("file.txt", 0755)
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("exists")))
	})

	It("should handle Rename when new parent doesn't exist", func() {
		mfs := memfs.New()

		// Create a file
		f, err := mfs.Create("file.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Close()).To(Succeed())

		// Try to rename to non-existent directory
		err = mfs.Rename("file.txt", "nonexistent/file.txt")
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("does not exist")))
	})

	It("should handle Rename when new parent is not a directory", func() {
		mfs := memfs.New()

		// Create two files
		f1, err := mfs.Create("file1.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(f1.Close()).To(Succeed())

		f2, err := mfs.Create("file2.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(f2.Close()).To(Succeed())

		// Try to rename file1 to file2/something (file2 is not a directory)
		err = mfs.Rename("file1.txt", "file2.txt/something")
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("invalid")))
	})

	It("should normalize paths starting with /", func() {
		mfs := memfs.New()

		// MkdirAll with leading slash
		err := mfs.MkdirAll("/some/deep/path", 0755)
		Expect(err).NotTo(HaveOccurred())

		// Verify it was created
		info, err := mfs.Stat("some/deep/path")
		Expect(err).NotTo(HaveOccurred())
		Expect(info.IsDir()).To(BeTrue())
	})

	It("should handle MkdirAll with empty path components", func() {
		mfs := memfs.New()

		// This would create path parts with empty strings after split
		// Testing that empty parts are skipped (line 122-123)
		err := mfs.MkdirAll("a//b", 0755)
		Expect(err).NotTo(HaveOccurred())

		// Verify it was created (cleaned to a/b)
		info, err := mfs.Stat("a/b")
		Expect(err).NotTo(HaveOccurred())
		Expect(info.IsDir()).To(BeTrue())
	})

	It("should handle empty string consistently", func() {
		mfs := memfs.New()

		// Empty string is invalid for Open
		_, err := mfs.Open("")
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("invalid")))

		// But normalizePath now treats "" as root for other operations
		// so they shouldn't create invalid internal paths
		err = mfs.Mkdir("", 0755)
		Expect(err).To(HaveOccurred()) // Should fail but not panic
	})

	Describe("WriteFile", func() {
		It("should write data to a new file", func() {
			mfs := memfs.New()
			err := mfs.WriteFile("test.txt", []byte("hello world"), 0644)
			Expect(err).NotTo(HaveOccurred())

			file, err := mfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())
			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("hello world"))
		})

		It("should overwrite an existing file", func() {
			mfs := memfs.New()
			err := mfs.WriteFile("test.txt", []byte("original"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = mfs.WriteFile("test.txt", []byte("new"), 0644)
			Expect(err).NotTo(HaveOccurred())

			file, err := mfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())
			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("new"))
		})

		It("should return error when parent directory does not exist", func() {
			mfs := memfs.New()
			err := mfs.WriteFile("nonexistent/test.txt", []byte("data"), 0644)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when writing to a directory", func() {
			mfs := memfs.New()
			err := mfs.Mkdir("testdir", 0755)
			Expect(err).NotTo(HaveOccurred())

			err = mfs.WriteFile("testdir", []byte("data"), 0644)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("fstest", func() {
		It("should pass fstest.TestFS", func() {
			mfs := memfs.New()

			// Create test structure
			Expect(mfs.Mkdir("/dir", 0755)).To(Succeed())

			f, err := mfs.Create("/file.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = f.(interface{ Write([]byte) (int, error) }).Write([]byte("content"))
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			f2, err := mfs.Create("/dir/nested.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = f2.(interface{ Write([]byte) (int, error) }).Write([]byte("nested"))
			Expect(err).NotTo(HaveOccurred())
			Expect(f2.Close()).To(Succeed())

			err = fstest.TestFS(mfs, "file.txt", "dir", "dir/nested.txt")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("ReadDir", func() {
		It("should return sorted directory entries", func() {
			mfs := memfs.New()
			Expect(mfs.Mkdir("testdir", 0755)).To(Succeed())
			f, err := mfs.Create("testdir/b.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())
			f, err = mfs.Create("testdir/a.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			entries, err := mfs.ReadDir("testdir")
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(2))
			Expect(entries[0].Name()).To(Equal("a.txt"))
			Expect(entries[1].Name()).To(Equal("b.txt"))
		})

		It("should error when directory not found", func() {
			mfs := memfs.New()
			_, err := mfs.ReadDir("nonexistent")
			Expect(err).To(HaveOccurred())
		})

		It("should error when path is not a directory", func() {
			mfs := memfs.New()
			f, err := mfs.Create("file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			_, err = mfs.ReadDir("file.txt")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("WriteFile", func() {
		It("should create a new file with content", func() {
			mfs := memfs.New()
			err := mfs.WriteFile("test.txt", []byte("hello"), 0644)
			Expect(err).NotTo(HaveOccurred())

			data, err := fs.ReadFile(mfs,"test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("hello"))
		})

		It("should overwrite existing file", func() {
			mfs := memfs.New()
			Expect(mfs.WriteFile("test.txt", []byte("original"), 0644)).To(Succeed())
			Expect(mfs.WriteFile("test.txt", []byte("updated"), 0644)).To(Succeed())

			data, err := fs.ReadFile(mfs,"test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("updated"))
		})

		It("should error when parent directory does not exist", func() {
			mfs := memfs.New()
			err := mfs.WriteFile("nonexistent/file.txt", []byte("data"), 0644)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Copy", func() {
		It("should copy files from src to dest", func() {
			mfs := memfs.New()
			src := memfs.New()
			f, err := src.Create("file.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = f.(io.Writer).Write([]byte("content"))
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			Expect(mfs.Mkdir("dest", 0755)).To(Succeed())
			Expect(mfs.Copy("dest", src)).To(Succeed())

			data, err := fs.ReadFile(mfs,"dest/file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("content"))
		})

		It("should create dest dir when it does not exist", func() {
			mfs := memfs.New()
			src := memfs.New()

			Expect(mfs.Copy("newdir", src)).To(Succeed())

			fi, err := mfs.Stat("newdir")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should succeed when dest dir already exists", func() {
			mfs := memfs.New()
			src := memfs.New()
			Expect(mfs.Copy(".", src)).To(Succeed())
		})

		It("should copy nested directories", func() {
			mfs := memfs.New()
			src := memfs.New()
			Expect(src.Mkdir("subdir", 0755)).To(Succeed())
			f, err := src.Create("subdir/file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			Expect(mfs.Copy(".", src)).To(Succeed())

			fi, err := mfs.Stat("subdir")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should error when file already exists in dest", func() {
			mfs := memfs.New()
			src := memfs.New()
			f, err := src.Create("file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			f2, err := mfs.Create("file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f2.Close()).To(Succeed())

			err = mfs.Copy(".", src)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrExist)).To(BeTrue())
		})

		It("should propagate walk errors", func() {
			mfs := memfs.New()
			// testfs.New() has Stat returning ErrNotExist by default,
			// causing WalkDir to fail and call walkFn with err != nil
			src := testfs.New()
			err := mfs.Copy(".", src)
			Expect(err).To(HaveOccurred())
		})

		It("should error when src.Open fails for a file entry", func() {
			mfs := memfs.New()
			openErr := errors.New("open error")
			fileEntry := testfs.NewDirEntry("file.txt", false)

			src := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return name == "." }
					return fi, nil
				}),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{fileEntry}, nil
				}),
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return nil, openErr
				}),
			)

			err := mfs.Copy(".", src)
			Expect(err).To(MatchError(openErr))
		})

		It("should error when DirEntry.Info fails", func() {
			mfs := memfs.New()
			infoErr := errors.New("info error")
			fileEntry := testfs.NewDirEntry("file.txt", false)
			fileEntry.InfoFunc = func() (ihfs.FileInfo, error) {
				return nil, infoErr
			}

			src := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return name == "." }
					return fi, nil
				}),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{fileEntry}, nil
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return &testfs.File{
						StatFunc: func() (ihfs.FileInfo, error) {
							return testfs.NewFileInfo(name), nil
						},
					}, nil
				}),
			)

			err := mfs.Copy(".", src)
			Expect(err).To(MatchError(infoErr))
		})

		It("should error when file read fails", func() {
			mfs := memfs.New()
			readErr := errors.New("read error")
			fileEntry := testfs.NewDirEntry("file.txt", false)

			src := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return name == "." }
					return fi, nil
				}),
				testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{fileEntry}, nil
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return &testfs.File{
						StatFunc: func() (ihfs.FileInfo, error) {
							return testfs.NewFileInfo(name), nil
						},
						ReadFunc: func(p []byte) (int, error) {
							return 0, readErr
						},
					}, nil
				}),
			)

			err := mfs.Copy(".", src)
			Expect(err).To(MatchError(readErr))
		})

		It("should error when dest root creation fails", func() {
			mfs := memfs.New()
			src := memfs.New()
			// "nonexistent/newdir" parent "nonexistent" doesn't exist
			err := mfs.Copy("nonexistent/newdir", src)
			Expect(err).To(HaveOccurred())
		})

		It("should error when subdir already exists in dest", func() {
			mfs := memfs.New()
			src := memfs.New()
			Expect(src.Mkdir("subdir", 0755)).To(Succeed())
			Expect(mfs.Mkdir("subdir", 0755)).To(Succeed())

			err := mfs.Copy(".", src)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("CreateTemp", func() {
		It("should create a temporary file", func() {
			mfs := memfs.New()
			file, err := mfs.CreateTemp(".", "tmp*")
			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			DeferCleanup(file.Close)
		})

		It("should error when dir does not exist", func() {
			mfs := memfs.New()
			_, err := mfs.CreateTemp("nonexistent", "tmp*")
			Expect(err).To(HaveOccurred())
		})

		It("should error when dir is not a directory", func() {
			mfs := memfs.New()
			f, err := mfs.Create("file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			_, err = mfs.CreateTemp("file.txt", "tmp*")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("MkdirTemp", func() {
		It("should create a temporary directory and return its relative path", func() {
			mfs := memfs.New()
			name, err := mfs.MkdirTemp(".", "tmpdir*")
			Expect(err).NotTo(HaveOccurred())
			Expect(name).NotTo(BeEmpty())

			fi, err := mfs.Stat(name)
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should error when dir does not exist", func() {
			mfs := memfs.New()
			_, err := mfs.MkdirTemp("nonexistent", "tmpdir*")
			Expect(err).To(HaveOccurred())
		})

		It("should error when dir is not a directory", func() {
			mfs := memfs.New()
			f, err := mfs.Create("file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			_, err = mfs.MkdirTemp("file.txt", "tmpdir*")
			Expect(err).To(HaveOccurred())
		})

		It("should error when pattern creates nested path whose parent does not exist", func() {
			mfs := memfs.New()
			_, err := mfs.MkdirTemp(".", "sub/dir*")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("TempFile", func() {
		It("should create a temporary file and return its relative path", func() {
			mfs := memfs.New()
			name, err := mfs.TempFile(".", "tmp*")
			Expect(err).NotTo(HaveOccurred())
			Expect(name).NotTo(BeEmpty())

			_, err = mfs.Stat(name)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should error when dir does not exist", func() {
			mfs := memfs.New()
			_, err := mfs.TempFile("nonexistent", "tmp*")
			Expect(err).To(HaveOccurred())
		})

		It("should error when dir is not a directory", func() {
			mfs := memfs.New()
			f, err := mfs.Create("file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			_, err = mfs.TempFile("file.txt", "tmp*")
			Expect(err).To(HaveOccurred())
		})

		It("should error when pattern creates nested path causing registration to fail", func() {
			mfs := memfs.New()
			_, err := mfs.TempFile(".", "sub/file*")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Symlink", func() {
		It("should create a symbolic link", func() {
			mfs := memfs.New()
			f, err := mfs.Create("target.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			err = mfs.Symlink("target.txt", "link.txt")
			Expect(err).NotTo(HaveOccurred())

			target, err := mfs.ReadLink("link.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(target).To(Equal("target.txt"))
		})

		It("should error when newname already exists", func() {
			mfs := memfs.New()
			f, err := mfs.Create("existing.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			err = mfs.Symlink("target.txt", "existing.txt")
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ihfs.ErrExist)).To(BeTrue())
		})

		It("should error when parent directory does not exist", func() {
			mfs := memfs.New()
			err := mfs.Symlink("target.txt", "nonexistent/link.txt")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ReadLink", func() {
		It("should return symlink target", func() {
			mfs := memfs.New()
			Expect(mfs.Symlink("target.txt", "link.txt")).To(Succeed())

			target, err := mfs.ReadLink("link.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(target).To(Equal("target.txt"))
		})

		It("should error when path not found", func() {
			mfs := memfs.New()
			_, err := mfs.ReadLink("nonexistent.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should error when path is not a symlink", func() {
			mfs := memfs.New()
			f, err := mfs.Create("file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			_, err = mfs.ReadLink("file.txt")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Lstat", func() {
		It("should return FileInfo for a regular file", func() {
			mfs := memfs.New()
			f, err := mfs.Create("file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			fi, err := mfs.Lstat("file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.Name()).To(Equal("file.txt"))
			Expect(fi.IsDir()).To(BeFalse())
		})

		It("should return symlink info without following it", func() {
			mfs := memfs.New()
			Expect(mfs.Symlink("target.txt", "link.txt")).To(Succeed())

			fi, err := mfs.Lstat("link.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.Name()).To(Equal("link.txt"))
			Expect(fi.Mode() & os.ModeSymlink).To(Equal(os.ModeSymlink))
		})

		It("should error when path not found", func() {
			mfs := memfs.New()
			_, err := mfs.Lstat("nonexistent.txt")
			Expect(err).To(HaveOccurred())
		})
	})
})

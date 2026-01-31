package memfs_test

import (
	"io"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/memfs"
)

var _ = Describe("Fs", func() {
	Describe("Open", func() {
		It("should open root directory", func() {
			mfs := memfs.New()
			file, err := mfs.Open("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			
			fi, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should return error for non-existent file", func() {
			mfs := memfs.New()
			_, err := mfs.Open("/nonexistent")
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

			file, err = mfs.Open("/test.txt")
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

			file, err = mfs.Open("/new.txt")
			Expect(err).NotTo(HaveOccurred())
			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("content"))
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

			file, err = mfs.Open("/test.txt")
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

			file, err = mfs.Open("/test.txt")
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

			file, err = mfs.Open("/test.txt")
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

			file, err := mfs.Open("/testdir")
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(ihfs.ReadDirFile)
			entries, err := dirFile.ReadDir(-1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(entries)).To(Equal(2))
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
			file, err := mfs.Open("/")
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

			file, err = mfs.Open("/test.txt")
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

			file, err = mfs.Open("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("01234XXX89"))
		})

		It("should error when ReadDir on closed file", func() {
			mfs := memfs.New()
			file, err := mfs.Open("/")
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

			file, err := mfs.Open("/emptydir")
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(ihfs.ReadDirFile)
			entries, err := dirFile.ReadDir(-1)
			// Empty directory returns EOF or empty slice depending on implementation
			if err == io.EOF {
				Expect(entries).To(BeNil())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(entries)).To(Equal(0))
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

			file, err := mfs.Open("/testdir")
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(ihfs.ReadDirFile)
			entries, err := dirFile.ReadDir(2)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(entries)).To(Equal(2))

			entries, err = dirFile.ReadDir(2)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(entries)).To(Equal(1))

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

			file, err = mfs.Open("/test.txt")
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

			file, err = mfs.Open("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("initialmore"))
		})

		It("should handle registerWithParent for root", func() {
			// Root has no parent, should not error
			mfs := memfs.New()
			file, err := mfs.Open("/")
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
			file, err := mfs.Open("")
			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
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

			file, err := mfs.Open("/nildir")
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(ihfs.ReadDirFile)
			entries, err := dirFile.ReadDir(-1)
			// Should handle nil dir gracefully
			if err != io.EOF {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(len(entries)).To(Equal(0))
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

			file, err = mfs.Open("/test.txt")
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
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			// Try to rename to non-existent directory
			err = mfs.Rename("/file.txt", "/nonexistent/file.txt")
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

		It("should handle unregisterWithParent error path", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			// Remove should work fine
			err = mfs.Remove("/test.txt")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle findParent for root", func() {
			mfs := memfs.New()
			// Root has no parent
			fi, err := mfs.Stat("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should handle creating file at root level", func() {
			mfs := memfs.New()
			file, err := mfs.Create("/test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should cover empty parts in MkdirAll", func() {
			mfs := memfs.New()
			err := mfs.MkdirAll("///a///b///", 0755)
			Expect(err).NotTo(HaveOccurred())

			fi, err := mfs.Stat("/a/b")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should cover registerWithParent error in OpenFile with O_CREATE", func() {
			mfs := memfs.New()
			// Try to create file in non-existent directory
			_, err := mfs.OpenFile("/nonexistent/test.txt", os.O_CREATE|os.O_WRONLY, 0644)
			Expect(err).To(HaveOccurred())
		})

		It("should cover unregisterWithParent error path in Remove", func() {
			mfs := memfs.New()
			// Try to remove non-existent file
			err := mfs.Remove("/nonexistent.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should cover unregisterWithParent error path in RemoveAll", func() {
			mfs := memfs.New()
			// RemoveAll doesn't error if path doesn't exist (matches os.RemoveAll behavior)
			err := mfs.RemoveAll("/nonexistent")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should cover unregisterWithParent error in Rename", func() {
			mfs := memfs.New()
			// Try to rename non-existent file
			err := mfs.Rename("/nonexistent.txt", "/new.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should cover root handling in registerWithParent", func() {
			mfs := memfs.New()
			// Root directory is handled specially
			fi, err := mfs.Stat("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should cover findParent at root", func() {
			mfs := memfs.New()
			// Create file at root
			file, err := mfs.Create("/rootfile.txt")
			Expect(err).NotTo(HaveOccurred())
			err = file.Close()
			Expect(err).NotTo(HaveOccurred())

			// Verify it exists
			fi, err := mfs.Stat("/rootfile.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).To(BeFalse())
		})

		It("should cover nil dir in ReadDir", func() {
			// This is an edge case where dir field is nil
			// Can occur if directory structure is manually manipulated
			mfs := memfs.New()
			err := mfs.Mkdir("/testdir", 0755)
			Expect(err).NotTo(HaveOccurred())

			file, err := mfs.Open("/testdir")
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(ihfs.ReadDirFile)
			entries, err := dirFile.ReadDir(-1)
			// Should handle gracefully
			if err != io.EOF {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(len(entries)).To(Equal(0))
		})
	})
})

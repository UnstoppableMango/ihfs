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
})

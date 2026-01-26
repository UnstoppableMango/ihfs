package testfs_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("FileInfo", func() {
	It("should create file info with name", func() {
		fi := testfs.FileInfo("test.txt")
		Expect(fi.Name()).To(Equal("test.txt"))
	})

	It("should return correct is directory status", func() {
		fi := testfs.FileInfo("test.txt", testfs.IsDir(true))
		Expect(fi.IsDir()).To(BeTrue())

		fi2 := testfs.FileInfo("file.txt", testfs.IsDir(false))
		Expect(fi2.IsDir()).To(BeFalse())
	})

	It("should return size", func() {
		fi := testfs.FileInfo("test.txt", testfs.WithSize(1024))
		Expect(fi.Size()).To(Equal(int64(1024)))
	})

	It("should return mode", func() {
		fi := testfs.FileInfo("test.txt", testfs.WithMode(0644))
		Expect(fi.Mode()).To(Equal(testfs.WithMode(0644).Mode))
	})

	It("should return mod time", func() {
		now := time.Now()
		fi := testfs.FileInfo("test.txt", testfs.WithModTime(now))
		Expect(fi.ModTime()).To(Equal(now))
	})

	It("should return sys", func() {
		fi := testfs.FileInfo("test.txt")
		Expect(fi.Sys()).To(BeNil())
	})

	Describe("Options", func() {
		It("should support IsDir option", func() {
			fi := testfs.FileInfo("dir", testfs.IsDir(true))
			Expect(fi.IsDir()).To(BeTrue())
		})

		It("should support WithSize option", func() {
			fi := testfs.FileInfo("file.txt", testfs.WithSize(512))
			Expect(fi.Size()).To(Equal(int64(512)))
		})

		It("should support WithMode option", func() {
			opt := testfs.WithMode(0755)
			fi := testfs.FileInfo("file.txt", opt)
			Expect(fi.Mode()).To(Equal(opt.Mode))
		})

		It("should support WithModTime option", func() {
			t := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			fi := testfs.FileInfo("file.txt", testfs.WithModTime(t))
			Expect(fi.ModTime()).To(Equal(t))
		})

		It("should support multiple options", func() {
			t := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			opt := testfs.WithMode(0644)
			fi := testfs.FileInfo("file.txt",
				testfs.IsDir(false),
				testfs.WithSize(2048),
				opt,
				testfs.WithModTime(t),
			)

			Expect(fi.Name()).To(Equal("file.txt"))
			Expect(fi.IsDir()).To(BeFalse())
			Expect(fi.Size()).To(Equal(int64(2048)))
			Expect(fi.Mode()).To(Equal(opt.Mode))
			Expect(fi.ModTime()).To(Equal(t))
		})
	})
})

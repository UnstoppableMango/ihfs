package testfs_test

import (
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("Fs", func() {
	Describe("New", func() {
		It("should create a new test filesystem", func() {
			tfs := testfs.New()
			Expect(tfs).NotTo(BeNil())
		})

		It("should accept options", func() {
			called := false
			openFunc := func(name string) (fs.File, error) {
				called = true
				return nil, fs.ErrNotExist
			}

			tfs := testfs.New(testfs.WithOpen(openFunc))
			Expect(tfs).NotTo(BeNil())

			_, err := tfs.Open("test.txt")
			Expect(err).To(MatchError(fs.ErrNotExist))
			Expect(called).To(BeTrue())
		})
	})

	Describe("Open", func() {
		It("should use custom open function when provided", func() {
			openFunc := func(name string) (fs.File, error) {
				return nil, fs.ErrPermission
			}

			tfs := testfs.New(testfs.WithOpen(openFunc))
			_, err := tfs.Open("test.txt")
			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("should use default open function when not provided", func() {
			tfs := testfs.New()
			_, err := tfs.Open("test.txt")
			Expect(err).To(MatchError(fs.ErrNotExist))
		})
	})

	Describe("Stat", func() {
		It("should use custom stat function when provided", func() {
			statFunc := func(name string) (fs.FileInfo, error) {
				return nil, fs.ErrPermission
			}

			tfs := testfs.New(testfs.WithStat(statFunc))
			_, err := tfs.Stat("test.txt")
			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("should use default stat function when not provided", func() {
			tfs := testfs.New()
			_, err := tfs.Stat("test.txt")
			Expect(err).To(MatchError(fs.ErrNotExist))
		})
	})

	Describe("WriteFile", func() {
		It("should use custom write function when provided", func() {
			writeFunc := func(name string, data []byte, perm fs.FileMode) error {
				return fs.ErrPermission
			}

			tfs := testfs.New(testfs.WithWriteFile(writeFunc))
			err := tfs.WriteFile("test.txt", []byte("data"), 0644)
			Expect(err).To(MatchError(fs.ErrPermission))
		})

		It("should use default write function when not provided", func() {
			tfs := testfs.New()
			err := tfs.WriteFile("test.txt", []byte("data"), 0644)
			Expect(err).To(MatchError(fs.ErrNotExist))
		})
	})

	Describe("Options", func() {
		It("should support WithOpen option", func() {
			openFunc := func(name string) (fs.File, error) {
				return nil, fs.ErrInvalid
			}

			tfs := testfs.New(testfs.WithOpen(openFunc))
			_, err := tfs.Open("test.txt")
			Expect(err).To(MatchError(fs.ErrInvalid))
		})

		It("should support WithStat option", func() {
			statFunc := func(name string) (fs.FileInfo, error) {
				return nil, fs.ErrInvalid
			}

			tfs := testfs.New(testfs.WithStat(statFunc))
			_, err := tfs.Stat("test.txt")
			Expect(err).To(MatchError(fs.ErrInvalid))
		})

		It("should support WithWriteFile option", func() {
			writeFunc := func(name string, data []byte, perm fs.FileMode) error {
				return fs.ErrInvalid
			}

			tfs := testfs.New(testfs.WithWriteFile(writeFunc))
			err := tfs.WriteFile("test.txt", []byte("data"), 0644)
			Expect(err).To(MatchError(fs.ErrInvalid))
		})
	})
})

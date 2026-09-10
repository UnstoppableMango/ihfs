package errfs_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/errfs"
)

var _ = Describe("Fs", func() {
	var (
		sentinel = errors.New("sentinel error")
		fsys     *errfs.Fs
	)

	BeforeEach(func() {
		fsys = errfs.New(sentinel)
	})

	It("should return the error from Open", func() {
		_, err := fsys.Open("file.txt")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Stat", func() {
		_, err := fsys.Stat("file.txt")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Create", func() {
		_, err := fsys.Create("file.txt")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from CreateTemp", func() {
		_, err := fsys.CreateTemp("", "prefix-*")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from WriteFile", func() {
		err := fsys.WriteFile("file.txt", []byte("data"), 0o644)
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from ReadFile", func() {
		_, err := fsys.ReadFile("file.txt")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Chmod", func() {
		err := fsys.Chmod("file.txt", 0o644)
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Chown", func() {
		err := fsys.Chown("file.txt", 0, 0)
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Chtimes", func() {
		err := fsys.Chtimes("file.txt", time.Time{}, time.Time{})
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Copy", func() {
		err := fsys.Copy("dir", errfs.New(nil))
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Glob", func() {
		_, err := fsys.Glob("*.txt")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Lstat", func() {
		_, err := fsys.Lstat("file.txt")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Mkdir", func() {
		err := fsys.Mkdir("dir", 0o755)
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from MkdirAll", func() {
		err := fsys.MkdirAll("path/to/dir", 0o755)
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from MkdirTemp", func() {
		_, err := fsys.MkdirTemp("", "prefix-*")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from OpenFile", func() {
		_, err := fsys.OpenFile("file.txt", 0, 0o644)
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from ReadDir", func() {
		_, err := fsys.ReadDir("dir")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from ReadDirNames", func() {
		_, err := fsys.ReadDirNames("dir")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from ReadLink", func() {
		_, err := fsys.ReadLink("symlink")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Remove", func() {
		err := fsys.Remove("file.txt")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from RemoveAll", func() {
		err := fsys.RemoveAll("dir")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Rename", func() {
		err := fsys.Rename("old.txt", "new.txt")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Sub", func() {
		_, err := fsys.Sub("subdir")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from Symlink", func() {
		err := fsys.Symlink("target", "link")
		Expect(err).To(MatchError(sentinel))
	})

	It("should return the error from TempFile", func() {
		_, err := fsys.TempFile("", "prefix-*")
		Expect(err).To(MatchError(sentinel))
	})

	It("should implement ihfs.FS", func() {
		var _ ihfs.FS = fsys
	})

	It("should implement ihfs.StatFS", func() {
		var _ ihfs.StatFS = fsys
	})

	It("should implement ihfs.CreateFS", func() {
		var _ ihfs.CreateFS = fsys
	})

	It("should implement ihfs.ReadDirFS", func() {
		var _ ihfs.ReadDirFS = fsys
	})
})

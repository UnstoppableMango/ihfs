package filter_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/filter"
	"github.com/unstoppablemango/ihfs/memfs"
	"github.com/unstoppablemango/ihfs/op"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("Ignore", func() {
	var fsys *ihfs.FilterFS

	BeforeEach(func() {
		base := memfs.New()
		fsys = ihfs.Filter(base, filter.Ignore([]string{"*.txt"}))
	})

	It("should block matching file via Open", func() {
		_, err := fsys.Open("blocked.txt")
		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should allow non-matching file via Open", func() {
		fn := filter.Ignore([]string{"*.txt"})
		base := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
			return &testfs.BoringFile{}, nil
		}))
		fsys := ihfs.Filter(base, fn)

		_, err := fsys.Open("main.go")

		Expect(err).NotTo(HaveOccurred())
	})

	It("should block matching file via Stat", func() {
		fn := filter.Ignore([]string{"blocked.txt"})
		base := memfs.New()
		fsys := ihfs.Filter(base, fn)

		_, err := fsys.Stat("blocked.txt")

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should block matching file via ReadDir op", func() {
		fn := filter.Ignore([]string{"*.txt"})
		base := memfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.ReadDir{Name: "notes.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should block matching file via Lstat op", func() {
		fn := filter.Ignore([]string{"*.txt"})
		base := memfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.Lstat{Name: "notes.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should block matching file via ReadFile op", func() {
		fn := filter.Ignore([]string{"*.txt"})
		base := memfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.ReadFile{Name: "notes.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should block matching file via ReadLink op", func() {
		fn := filter.Ignore([]string{"*.txt"})
		base := memfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.ReadLink{Name: "notes.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should block matching file via WriteFile op", func() {
		fn := filter.Ignore([]string{"*.txt"})
		base := memfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.WriteFile{Name: "notes.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should block matching file via Remove op", func() {
		fn := filter.Ignore([]string{"*.txt"})
		base := memfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.Remove{Name: "notes.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should block matching file via RemoveAll op", func() {
		fn := filter.Ignore([]string{"*.txt"})
		base := memfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.RemoveAll{Name: "notes.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should pass through op.Glob (default case)", func() {
		fn := filter.Ignore([]string{"*.txt"})
		base := memfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.Glob{Pattern: "*.txt"})

		Expect(err).NotTo(HaveOccurred())
	})

	It("should block files under a matched directory", func() {
		fn := filter.Ignore([]string{"vendor"})
		base := memfs.New()
		fsys := ihfs.Filter(base, fn)

		_, err := fsys.Open("vendor/pkg/file.go")

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should allow directories regardless of pattern", func() {
		fn := filter.Ignore([]string{"*.txt"})
		dirInfo := testfs.NewFileInfo("notes.txt")
		dirInfo.IsDirFunc = func() bool { return true }
		base := testfs.New(testfs.WithStat(func(string) (ihfs.FileInfo, error) {
			return dirInfo, nil
		}))
		fsys := ihfs.Filter(base, fn)

		result, err := fsys.Stat("notes.txt")

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(dirInfo))
	})
})

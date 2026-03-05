package filter_test

import (
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/filter"
	"github.com/unstoppablemango/ihfs/op"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("NameRegex", func() {
	var re *regexp.Regexp

	BeforeEach(func() {
		re = regexp.MustCompile(`\.go$`)
	})

	It("should allow matching file via Open", func() {
		file := &testfs.BoringFile{}
		base := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
			return file, nil
		}))
		fsys := ihfs.Filter(base, filter.NameRegex(re))

		f, err := fsys.Open("main.go")

		Expect(err).NotTo(HaveOccurred())
		Expect(f).To(BeIdenticalTo(file))
	})

	It("should return ErrPermission for non-matching name via Open", func() {
		base := testfs.New()
		fsys := ihfs.Filter(base, filter.NameRegex(re))

		_, err := fsys.Open("main.txt")

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should allow matching file via Stat", func() {
		info := testfs.NewFileInfo("main.go")
		base := testfs.New(testfs.WithStat(func(string) (ihfs.FileInfo, error) {
			return info, nil
		}))
		fsys := ihfs.Filter(base, filter.NameRegex(re))

		result, err := fsys.Stat("main.go")

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(info))
	})

	It("should return ErrPermission for non-matching name via Stat", func() {
		base := testfs.New()
		fsys := ihfs.Filter(base, filter.NameRegex(re))

		_, err := fsys.Stat("main.txt")

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should allow directories regardless of regexp", func() {
		dirInfo := testfs.NewFileInfo("somedir")
		dirInfo.IsDirFunc = func() bool { return true }
		base := testfs.New(testfs.WithStat(func(string) (ihfs.FileInfo, error) {
			return dirInfo, nil
		}))
		fsys := ihfs.Filter(base, filter.NameRegex(re))

		result, err := fsys.Stat("somedir")

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeIdenticalTo(dirInfo))
	})

	It("should pass through op.Glob (default case)", func() {
		fn := filter.NameRegex(re)
		base := testfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.Glob{Pattern: "*.txt"})

		Expect(err).NotTo(HaveOccurred())
	})

	It("should filter op.ReadDir by name", func() {
		fn := filter.NameRegex(re)
		base := testfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.ReadDir{Name: "somedir.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should filter op.Lstat by name", func() {
		fn := filter.NameRegex(re)
		base := testfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.Lstat{Name: "main.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should filter op.ReadFile by name", func() {
		fn := filter.NameRegex(re)
		base := testfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.ReadFile{Name: "main.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should filter op.ReadLink by name", func() {
		fn := filter.NameRegex(re)
		base := testfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.ReadLink{Name: "main.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should filter op.WriteFile by name", func() {
		fn := filter.NameRegex(re)
		base := testfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.WriteFile{Name: "main.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should filter op.Remove by name", func() {
		fn := filter.NameRegex(re)
		base := testfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.Remove{Name: "main.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})

	It("should filter op.RemoveAll by name", func() {
		fn := filter.NameRegex(re)
		base := testfs.New()
		fsys := ihfs.Filter(base, fn)

		err := fn(fsys, op.RemoveAll{Name: "main.txt"})

		Expect(err).To(MatchError(ihfs.ErrPermission))
	})
})

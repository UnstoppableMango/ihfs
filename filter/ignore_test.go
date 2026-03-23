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
	Describe("op dispatch", func() {
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

	Describe("pattern parsing", func() {
		It("should ignore blank lines", func() {
			fn := filter.Ignore([]string{"", "  ", "\t"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			err := fn(fsys, op.Open{Name: "anything.txt"})

			Expect(err).NotTo(HaveOccurred())
		})

		It("should ignore comment lines starting with #", func() {
			fn := filter.Ignore([]string{"# this is a comment", "#*.go"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			err := fn(fsys, op.Open{Name: "main.go"})

			Expect(err).NotTo(HaveOccurred())
		})

		It("should ignore trailing spaces and tabs", func() {
			fn := filter.Ignore([]string{"*.txt   \t"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			err := fn(fsys, op.Open{Name: "notes.txt"})

			Expect(err).To(MatchError(ihfs.ErrPermission))
		})

		It("should treat \\# as literal # (not a comment)", func() {
			fn := filter.Ignore([]string{`\#readme`})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			err := fn(fsys, op.Open{Name: "#readme"})

			Expect(err).To(MatchError(ihfs.ErrPermission))
		})

		It("should treat \\! as literal ! (not a negation)", func() {
			fn := filter.Ignore([]string{`\!important.txt`})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			err := fn(fsys, op.Open{Name: "!important.txt"})

			Expect(err).To(MatchError(ihfs.ErrPermission))
		})

		It("should skip lone ! (empty pattern after negation prefix)", func() {
			fn := filter.Ignore([]string{"!"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			err := fn(fsys, op.Open{Name: "anything.txt"})

			Expect(err).NotTo(HaveOccurred())
		})

		It("should skip lone / (empty pattern after dir-only suffix)", func() {
			fn := filter.Ignore([]string{"/"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			err := fn(fsys, op.Open{Name: "anything.txt"})

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("gitignore pattern semantics", func() {
		It("should match basename at any depth (no / in pattern)", func() {
			fn := filter.Ignore([]string{"*.log"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			Expect(fn(fsys, op.Open{Name: "app.log"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "logs/app.log"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "a/b/c/app.log"})).To(MatchError(ihfs.ErrPermission))
		})

		It("should not match non-matching basename", func() {
			fn := filter.Ignore([]string{"*.log"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			err := fn(fsys, op.Open{Name: "logs/app.txt"})

			Expect(err).NotTo(HaveOccurred())
		})

		It("should anchor pattern with / to root", func() {
			fn := filter.Ignore([]string{"src/*.go"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			Expect(fn(fsys, op.Open{Name: "src/main.go"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "other/main.go"})).NotTo(HaveOccurred())
			Expect(fn(fsys, op.Open{Name: "src/sub/main.go"})).NotTo(HaveOccurred())
		})

		It("should anchor leading / pattern to root", func() {
			fn := filter.Ignore([]string{"/vendor"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			Expect(fn(fsys, op.Open{Name: "vendor"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "pkg/vendor"})).NotTo(HaveOccurred())
		})

		It("should support ? wildcard", func() {
			fn := filter.Ignore([]string{"file?.txt"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			Expect(fn(fsys, op.Open{Name: "file1.txt"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "fileab.txt"})).NotTo(HaveOccurred())
		})

		It("should support [...] character class", func() {
			fn := filter.Ignore([]string{"file[0-9].txt"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			Expect(fn(fsys, op.Open{Name: "file3.txt"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "filea.txt"})).NotTo(HaveOccurred())
		})

		It("should support **/foo matching foo at any depth", func() {
			fn := filter.Ignore([]string{"**/secret.txt"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			Expect(fn(fsys, op.Open{Name: "secret.txt"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "a/secret.txt"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "a/b/secret.txt"})).To(MatchError(ihfs.ErrPermission))
		})

		It("should support foo/** matching everything inside foo", func() {
			fn := filter.Ignore([]string{"vendor/**"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			Expect(fn(fsys, op.Open{Name: "vendor/pkg/file.go"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "src/file.go"})).NotTo(HaveOccurred())
		})

		It("should support foo/**/bar matching bar at any depth inside foo", func() {
			fn := filter.Ignore([]string{"src/**/test.go"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			Expect(fn(fsys, op.Open{Name: "src/test.go"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "src/pkg/test.go"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "src/a/b/test.go"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "other/test.go"})).NotTo(HaveOccurred())
		})

		It("should not match when ** cannot satisfy remaining pattern segments", func() {
			fn := filter.Ignore([]string{"**/secret.txt"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			err := fn(fsys, op.Open{Name: "a/b/other.txt"})

			Expect(err).NotTo(HaveOccurred())
		})

		It("should not match rooted pattern when path is shorter than pattern", func() {
			fn := filter.Ignore([]string{"src/test.go"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			// "src" alone does not match "src/test.go" pattern
			err := fn(fsys, op.Open{Name: "src"})

			Expect(err).NotTo(HaveOccurred())
		})

		It("should skip dir-only patterns (ending with /)", func() {
			fn := filter.Ignore([]string{"vendor/"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			// dirOnly patterns are skipped; files named vendor are not blocked
			err := fn(fsys, op.Open{Name: "vendor"})

			Expect(err).NotTo(HaveOccurred())
		})

		It("should support negation to re-include a file", func() {
			fn := filter.Ignore([]string{"*.txt", "!important.txt"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			Expect(fn(fsys, op.Open{Name: "notes.txt"})).To(MatchError(ihfs.ErrPermission))
			Expect(fn(fsys, op.Open{Name: "important.txt"})).NotTo(HaveOccurred())
		})

		It("should apply patterns in order (last match wins)", func() {
			fn := filter.Ignore([]string{"!important.txt", "*.txt"})
			base := memfs.New()
			fsys := ihfs.Filter(base, fn)

			// *.txt comes after !important.txt so important.txt is still blocked
			err := fn(fsys, op.Open{Name: "important.txt"})

			Expect(err).To(MatchError(ihfs.ErrPermission))
		})
	})
})

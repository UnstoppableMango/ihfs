package ignore_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/filter/ignore"
)

var _ = Describe("Parse", func() {
	It("should return nil for a blank line", func() {
		Expect(ignore.Parse("")).To(BeNil())
	})

	It("should return nil for a whitespace-only line", func() {
		Expect(ignore.Parse("  ")).To(BeNil())
		Expect(ignore.Parse("\t")).To(BeNil())
	})

	It("should return nil for a comment line", func() {
		Expect(ignore.Parse("# this is a comment")).To(BeNil())
		Expect(ignore.Parse("#*.go")).To(BeNil())
	})

	It("should return nil for a lone !", func() {
		Expect(ignore.Parse("!")).To(BeNil())
	})

	It("should return nil for a lone /", func() {
		Expect(ignore.Parse("/")).To(BeNil())
	})

	It("should strip trailing spaces and tabs", func() {
		p := ignore.Parse("*.txt   \t")

		Expect(p).NotTo(BeNil())
		Expect(p.Match("notes.txt")).To(HaveValue(BeTrue()))
	})

	It("should treat \\# as a literal # (not a comment)", func() {
		p := ignore.Parse(`\#readme`)

		Expect(p).NotTo(BeNil())
		Expect(p.Match("#readme")).To(HaveValue(BeTrue()))
	})

	It("should treat \\! as a literal ! (not a negation)", func() {
		p := ignore.Parse(`\!important.txt`)

		Expect(p).NotTo(BeNil())
		Expect(p.Match("!important.txt")).To(HaveValue(BeTrue()))
	})
})

var _ = Describe("Pattern.Match", func() {
	It("should return nil for a nil pattern", func() {
		var p *ignore.Pattern
		Expect(p.Match("anything")).To(BeNil())
	})

	It("should return nil for a dir-only pattern", func() {
		p := ignore.Parse("vendor/")

		Expect(p.Match("vendor")).To(BeNil())
	})

	It("should match the base name at any depth (no / in pattern)", func() {
		p := ignore.Parse("*.log")

		Expect(p.Match("app.log")).To(HaveValue(BeTrue()))
		Expect(p.Match("logs/app.log")).To(HaveValue(BeTrue()))
		Expect(p.Match("a/b/c/app.log")).To(HaveValue(BeTrue()))
	})

	It("should return nil for a non-matching base name", func() {
		p := ignore.Parse("*.log")

		Expect(p.Match("logs/app.txt")).To(BeNil())
	})

	It("should anchor a pattern containing / to the root", func() {
		p := ignore.Parse("src/*.go")

		Expect(p.Match("src/main.go")).To(HaveValue(BeTrue()))
		Expect(p.Match("other/main.go")).To(BeNil())
		Expect(p.Match("src/sub/main.go")).To(BeNil())
	})

	It("should anchor a leading / pattern to the root", func() {
		p := ignore.Parse("/vendor")

		Expect(p.Match("vendor")).To(HaveValue(BeTrue()))
		Expect(p.Match("pkg/vendor")).To(BeNil())
	})

	It("should support ? wildcard", func() {
		p := ignore.Parse("file?.txt")

		Expect(p.Match("file1.txt")).To(HaveValue(BeTrue()))
		Expect(p.Match("fileab.txt")).To(BeNil())
	})

	It("should support [...] character class", func() {
		p := ignore.Parse("file[0-9].txt")

		Expect(p.Match("file3.txt")).To(HaveValue(BeTrue()))
		Expect(p.Match("filea.txt")).To(BeNil())
	})

	It("should support **/foo matching foo at any depth", func() {
		p := ignore.Parse("**/secret.txt")

		Expect(p.Match("secret.txt")).To(HaveValue(BeTrue()))
		Expect(p.Match("a/secret.txt")).To(HaveValue(BeTrue()))
		Expect(p.Match("a/b/secret.txt")).To(HaveValue(BeTrue()))
		Expect(p.Match("a/b/other.txt")).To(BeNil())
	})

	It("should support foo/** matching everything inside foo", func() {
		p := ignore.Parse("vendor/**")

		Expect(p.Match("vendor/pkg/file.go")).To(HaveValue(BeTrue()))
		Expect(p.Match("src/file.go")).To(BeNil())
	})

	It("should support foo/**/bar matching bar at any depth inside foo", func() {
		p := ignore.Parse("src/**/test.go")

		Expect(p.Match("src/test.go")).To(HaveValue(BeTrue()))
		Expect(p.Match("src/pkg/test.go")).To(HaveValue(BeTrue()))
		Expect(p.Match("src/a/b/test.go")).To(HaveValue(BeTrue()))
		Expect(p.Match("other/test.go")).To(BeNil())
	})

	It("should return nil when ** cannot satisfy remaining pattern segments", func() {
		p := ignore.Parse("**/secret.txt")

		Expect(p.Match("a/b/other.txt")).To(BeNil())
	})

	It("should return nil when path is shorter than a rooted pattern", func() {
		p := ignore.Parse("src/test.go")

		Expect(p.Match("src")).To(BeNil())
	})

	It("should return a pointer to false for a negated match", func() {
		p := ignore.Parse("!important.txt")

		Expect(p.Match("important.txt")).To(HaveValue(BeFalse()))
		Expect(p.Match("other.txt")).To(BeNil())
	})
})

var _ = Describe("Ignored", func() {
	It("should return false when no patterns match", func() {
		patterns := []ignore.Pattern{*ignore.Parse("*.log")}

		Expect(ignore.Ignored(patterns, "notes.txt")).To(BeFalse())
	})

	It("should return true when a pattern matches", func() {
		patterns := []ignore.Pattern{*ignore.Parse("*.txt")}

		Expect(ignore.Ignored(patterns, "notes.txt")).To(BeTrue())
	})

	It("should skip dir-only patterns", func() {
		patterns := []ignore.Pattern{*ignore.Parse("vendor/")}

		Expect(ignore.Ignored(patterns, "vendor")).To(BeFalse())
	})

	It("should support negation to re-include a file", func() {
		patterns := []ignore.Pattern{
			*ignore.Parse("*.txt"),
			*ignore.Parse("!important.txt"),
		}

		Expect(ignore.Ignored(patterns, "notes.txt")).To(BeTrue())
		Expect(ignore.Ignored(patterns, "important.txt")).To(BeFalse())
	})

	It("should apply patterns in order (last match wins)", func() {
		patterns := []ignore.Pattern{
			*ignore.Parse("!important.txt"),
			*ignore.Parse("*.txt"),
		}

		Expect(ignore.Ignored(patterns, "important.txt")).To(BeTrue())
	})
})

var _ = Describe("File", func() {
	var f ignore.File

	BeforeEach(func() {
		f = ignore.File{
			*ignore.Parse("*.log"),
			*ignore.Parse("!keep.log"),
		}
	})

	Describe("Patterns", func() {
		It("should return all patterns", func() {
			Expect(f.Patterns()).To(HaveLen(2))
		})
	})

	Describe("Iter", func() {
		It("should iterate over all patterns", func() {
			var count int
			for range f.Iter() {
				count++
			}
			Expect(count).To(Equal(2))
		})
	})

	Describe("Ignored", func() {
		It("should return true for a matching file", func() {
			Expect(f.Ignored("app.log")).To(BeTrue())
		})

		It("should return false for a negated file", func() {
			Expect(f.Ignored("keep.log")).To(BeFalse())
		})

		It("should return false for a non-matching file", func() {
			Expect(f.Ignored("main.go")).To(BeFalse())
		})

		It("should skip dir-only patterns", func() {
			f := ignore.File{*ignore.Parse("vendor/")}
			Expect(f.Ignored("vendor")).To(BeFalse())
		})
	})
})

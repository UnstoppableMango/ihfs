package ghfs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/ghfs"
)

// These test some invalid URLs. For example raw.githubusercontent.com/ does not
// follow the owner/repo/tree/branch pattern, and github.com/ does not follow
// the owner/repo/blob/branch/path pattern. I don't think that matters though,
// as the URL prefix stripping is a convenience feature.

var _ = Describe("Fs", func() {
	DescribeTableSubtree("Open",
		Entry(nil, "https://api.github.com/"),
		Entry(nil, "https://github.com/"),
		Entry(nil, "https://raw.githubusercontent.com/"),
		Entry(nil, "github.com/"),
		Entry(nil, "api.github.com/"),
		Entry(nil, "raw.githubusercontent.com/"),
		Entry("No prefix", ""),
		func(prefix string) {
			It("should parse an owner path", func() {
				fsys := ghfs.New()

				f, err := fsys.Open(prefix + "UnstoppableMango")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Owner{}))
				o := f.(*ghfs.Owner)
				Expect(o.Name()).To(Equal("UnstoppableMango"))
			})

			It("should parse a repository path", func() {
				fsys := ghfs.New()

				f, err := fsys.Open(prefix + "UnstoppableMango/ihfs")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Repository{}))
				o := f.(*ghfs.Repository)
				Expect(o.Owner()).To(Equal("UnstoppableMango"))
				Expect(o.Name()).To(Equal("ihfs"))
			})

			DescribeTable("should parse a release path",
				func(path string) {
					fsys := ghfs.New()

					f, err := fsys.Open(prefix + path)

					Expect(err).NotTo(HaveOccurred())
					Expect(f).To(BeAssignableToTypeOf(&ghfs.Release{}))
					o := f.(*ghfs.Release)
					Expect(o.Owner()).To(Equal("UnstoppableMango"))
					Expect(o.Repository()).To(Equal("ihfs"))
					Expect(o.Name()).To(Equal("v0.1.0"))
				},
				Entry(nil, "UnstoppableMango/ihfs/releases/tag/v0.1.0"),
				Entry(nil, "UnstoppableMango/ihfs/releases/download/v0.1.0"),
			)

			DescribeTable("should parse an asset path",
				func(path string) {
					fsys := ghfs.New()

					f, err := fsys.Open(prefix + path)

					Expect(err).NotTo(HaveOccurred())
					Expect(f).To(BeAssignableToTypeOf(&ghfs.Asset{}))
					o := f.(*ghfs.Asset)
					Expect(o.Owner()).To(Equal("UnstoppableMango"))
					Expect(o.Repository()).To(Equal("ihfs"))
					Expect(o.Release()).To(Equal("v0.1.0"))
					Expect(o.Name()).To(Equal("asset.tar.gz"))
				},
				Entry(nil, "UnstoppableMango/ihfs/releases/tag/v0.1.0/asset.tar.gz"),
				Entry(nil, "UnstoppableMango/ihfs/releases/download/v0.1.0/asset.tar.gz"),
			)

			It("should parse a branch path", func() {
				fsys := ghfs.New()

				f, err := fsys.Open(prefix + "UnstoppableMango/ihfs/tree/main")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Branch{}))
				o := f.(*ghfs.Branch)
				Expect(o.Owner()).To(Equal("UnstoppableMango"))
				Expect(o.Repository()).To(Equal("ihfs"))
				Expect(o.Name()).To(Equal("main"))
			})

			It("should parse a content path", func() {
				fsys := ghfs.New()

				f, err := fsys.Open(prefix + "UnstoppableMango/ihfs/blob/main/README.md")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Content{}))
				o := f.(*ghfs.Content)
				Expect(o.Owner()).To(Equal("UnstoppableMango"))
				Expect(o.Repository()).To(Equal("ihfs"))
				Expect(o.Branch()).To(Equal("main"))
				Expect(o.Name()).To(Equal("README.md"))
			})

			It("should parse a nested content path", func() {
				fsys := ghfs.New()

				f, err := fsys.Open(prefix + "UnstoppableMango/ihfs/blob/main/.github/renovate.json")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Content{}))
				o := f.(*ghfs.Content)
				Expect(o.Owner()).To(Equal("UnstoppableMango"))
				Expect(o.Repository()).To(Equal("ihfs"))
				Expect(o.Branch()).To(Equal("main"))
				Expect(o.Name()).To(Equal(".github/renovate.json"))
			})
		},
	)
})

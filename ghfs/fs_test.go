package ghfs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/ghfs"
)

var _ = Describe("Fs", func() {
	Describe("Open", func() {
		DescribeTable("should parse an owner path",
			func(path string) {
				fsys := ghfs.New()

				f, err := fsys.Open(path)

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Owner{}))
				o := f.(*ghfs.Owner)
				Expect(o.Name()).To(Equal("UnstoppableMango"))
			},
			Entry(nil, "https://api.github.com/UnstoppableMango"),
			Entry(nil, "https://github.com/UnstoppableMango"),
			Entry(nil, "github.com/UnstoppableMango"),
			Entry(nil, "api.github.com/UnstoppableMango"),
			Entry(nil, "raw.githubusercontent.com/UnstoppableMango"),
			Entry(nil, "UnstoppableMango"),
		)

		DescribeTable("should parse a repository path",
			func(path string) {
				fsys := ghfs.New()

				f, err := fsys.Open(path)

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Repository{}))
				o := f.(*ghfs.Repository)
				Expect(o.Owner()).To(Equal("UnstoppableMango"))
				Expect(o.Name()).To(Equal("ihfs"))
			},
			Entry(nil, "https://api.github.com/UnstoppableMango/ihfs"),
			Entry(nil, "https://github.com/UnstoppableMango/ihfs"),
			Entry(nil, "github.com/UnstoppableMango/ihfs"),
			Entry(nil, "api.github.com/UnstoppableMango/ihfs"),
			Entry(nil, "raw.githubusercontent.com/UnstoppableMango/ihfs"),
			Entry(nil, "UnstoppableMango/ihfs"),
		)

		DescribeTable("should parse a release path",
			func(path string) {
				fsys := ghfs.New()

				f, err := fsys.Open(path)

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Release{}))
				o := f.(*ghfs.Release)
				Expect(o.Owner()).To(Equal("UnstoppableMango"))
				Expect(o.Repository()).To(Equal("ihfs"))
				Expect(o.Name()).To(Equal("v0.1.0"))
			},
			Entry(nil, "https://api.github.com/UnstoppableMango/ihfs/releases/tag/v0.1.0"),
			Entry(nil, "https://github.com/UnstoppableMango/ihfs/releases/tag/v0.1.0"),
			Entry(nil, "github.com/UnstoppableMango/ihfs/releases/tag/v0.1.0"),
			Entry(nil, "api.github.com/UnstoppableMango/ihfs/releases/tag/v0.1.0"),
			Entry(nil, "raw.githubusercontent.com/UnstoppableMango/ihfs/releases/tag/v0.1.0"),
			Entry(nil, "UnstoppableMango/ihfs/releases/tag/v0.1.0"),
			Entry(nil, "https://api.github.com/UnstoppableMango/ihfs/releases/download/v0.1.0"),
			Entry(nil, "https://github.com/UnstoppableMango/ihfs/releases/download/v0.1.0"),
			Entry(nil, "github.com/UnstoppableMango/ihfs/releases/download/v0.1.0"),
			Entry(nil, "api.github.com/UnstoppableMango/ihfs/releases/download/v0.1.0"),
			Entry(nil, "raw.githubusercontent.com/UnstoppableMango/ihfs/releases/download/v0.1.0"),
			Entry(nil, "UnstoppableMango/ihfs/releases/download/v0.1.0"),
		)

		DescribeTable("should parse an asset path",
			func(path string) {
				fsys := ghfs.New()

				f, err := fsys.Open(path)

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Asset{}))
				o := f.(*ghfs.Asset)
				Expect(o.Owner()).To(Equal("UnstoppableMango"))
				Expect(o.Repository()).To(Equal("ihfs"))
				Expect(o.Release()).To(Equal("v0.1.0"))
				Expect(o.Name()).To(Equal("asset.tar.gz"))
			},
			Entry(nil, "https://api.github.com/UnstoppableMango/ihfs/releases/tag/v0.1.0/asset.tar.gz"),
			Entry(nil, "https://github.com/UnstoppableMango/ihfs/releases/tag/v0.1.0/asset.tar.gz"),
			Entry(nil, "github.com/UnstoppableMango/ihfs/releases/tag/v0.1.0/asset.tar.gz"),
			Entry(nil, "api.github.com/UnstoppableMango/ihfs/releases/tag/v0.1.0/asset.tar.gz"),
			Entry(nil, "raw.githubusercontent.com/UnstoppableMango/ihfs/releases/tag/v0.1.0/asset.tar.gz"),
			Entry(nil, "UnstoppableMango/ihfs/releases/tag/v0.1.0/asset.tar.gz"),
			Entry(nil, "https://api.github.com/UnstoppableMango/ihfs/releases/download/v0.1.0/asset.tar.gz"),
			Entry(nil, "https://github.com/UnstoppableMango/ihfs/releases/download/v0.1.0/asset.tar.gz"),
			Entry(nil, "github.com/UnstoppableMango/ihfs/releases/download/v0.1.0/asset.tar.gz"),
			Entry(nil, "api.github.com/UnstoppableMango/ihfs/releases/download/v0.1.0/asset.tar.gz"),
			Entry(nil, "raw.githubusercontent.com/UnstoppableMango/ihfs/releases/download/v0.1.0/asset.tar.gz"),
			Entry(nil, "UnstoppableMango/ihfs/releases/download/v0.1.0/asset.tar.gz"),
		)
	})
})

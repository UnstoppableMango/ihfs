package ghfs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-github/v73/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/unstoppablemango/ihfs"
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
			var fsys ihfs.FS

			BeforeEach(func() {
				mockHttp, s := mock.NewMockedHTTPClientAndServer(
					mock.WithRequestMatch(
						mock.GetUsersByUsername,
						github.User{
							Name: github.Ptr("test-user"),
						},
					),
					mock.WithRequestMatch(
						mock.GetReposByOwnerByRepo,
						github.Repository{
							Name: github.Ptr("ihfs"),
							Owner: &github.User{
								Name: github.Ptr("test-user"),
							},
						},
					),
					mock.WithRequestMatch(
						mock.GetReposReleasesByOwnerByRepoByReleaseId,
						github.RepositoryRelease{
							Name: github.Ptr("test-release"),
						},
					),
					mock.WithRequestMatch(
						mock.GetReposReleasesAssetsByOwnerByRepoByAssetId,
						github.ReleaseAsset{
							Name: github.Ptr("asset.tar.gz"),
						},
					),
					mock.WithRequestMatch(
						mock.GetReposBranchesByOwnerByRepoByBranch,
						github.Branch{
							Name: github.Ptr("test-branch"),
						},
					),
					mock.WithRequestMatch(
						mock.GetReposContentsByOwnerByRepoByPath,
						github.RepositoryContent{
							Name: github.Ptr("file.txt"),
						},
					),
				)

				DeferCleanup(s.Close)
				fsys = ghfs.New(ghfs.WithHttpClient(mockHttp))
			})

			It("should parse an owner path", func() {
				f, err := fsys.Open(prefix + "test-user")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Owner{}))
				o := f.(*ghfs.Owner)
				Expect(o.Name()).To(Equal("test-user"))
				u, err := o.User()
				Expect(err).NotTo(HaveOccurred())
				Expect(u.GetName()).To(Equal("test-user"))
			})

			It("should parse a repository path", func() {
				f, err := fsys.Open(prefix + "test-user/test-repo")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Repository{}))
				r := f.(*ghfs.Repository)
				Expect(r.Owner()).To(Equal("test-user"))
				Expect(r.Name()).To(Equal("test-repo"))
				repo, err := r.Repository()
				Expect(err).NotTo(HaveOccurred())
				Expect(repo.GetName()).To(Equal("ihfs"))
				Expect(repo.GetOwner().GetName()).To(Equal("test-user"))
			})

			DescribeTable("should parse a release path",
				func(path string) {
					f, err := fsys.Open(prefix + path)

					Expect(err).NotTo(HaveOccurred())
					Expect(f).To(BeAssignableToTypeOf(&ghfs.Release{}))
					r := f.(*ghfs.Release)
					Expect(r.Owner()).To(Equal("test-user"))
					Expect(r.Repository()).To(Equal("test-repo"))
					Expect(r.Name()).To(Equal("v0.1.0"))
					release, err := r.Release()
					Expect(err).NotTo(HaveOccurred())
					Expect(release.GetName()).To(Equal("test-release"))
				},
				Entry(nil, "test-user/test-repo/releases/tag/v0.1.0"),
				Entry(nil, "test-user/test-repo/releases/download/v0.1.0"),
			)

			DescribeTable("should parse an asset path",
				func(path string) {
					f, err := fsys.Open(prefix + path)

					Expect(err).NotTo(HaveOccurred())
					Expect(f).To(BeAssignableToTypeOf(&ghfs.Asset{}))
					a := f.(*ghfs.Asset)
					Expect(a.Owner()).To(Equal("test-user"))
					Expect(a.Repository()).To(Equal("test-repo"))
					Expect(a.Release()).To(Equal("v0.1.0"))
					Expect(a.Name()).To(Equal("asset.tar.gz"))
					asset, err := a.Asset()
					Expect(err).NotTo(HaveOccurred())
					Expect(asset.GetName()).To(Equal("asset.tar.gz"))
				},
				Entry(nil, "test-user/test-repo/releases/tag/v0.1.0/asset.tar.gz"),
				Entry(nil, "test-user/test-repo/releases/download/v0.1.0/asset.tar.gz"),
			)

			It("should parse a branch path", func() {
				f, err := fsys.Open(prefix + "test-user/test-repo/tree/test-branch")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Branch{}))
				b := f.(*ghfs.Branch)
				Expect(b.Owner()).To(Equal("test-user"))
				Expect(b.Repository()).To(Equal("test-repo"))
				Expect(b.Name()).To(Equal("test-branch"))
				branch, err := b.Branch()
				Expect(err).NotTo(HaveOccurred())
				Expect(branch.GetName()).To(Equal("test-branch"))
			})

			It("should parse a content path", func() {
				f, err := fsys.Open(prefix + "test-user/test-repo/blob/test-branch/file.txt")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Content{}))
				c := f.(*ghfs.Content)
				Expect(c.Owner()).To(Equal("test-user"))
				Expect(c.Repository()).To(Equal("test-repo"))
				Expect(c.Branch()).To(Equal("test-branch"))
				Expect(c.Name()).To(Equal("file.txt"))
				content, err := c.Content()
				Expect(err).NotTo(HaveOccurred())
				Expect(content.GetName()).To(Equal("file.txt"))
			})

			It("should parse a nested content path", func() {
				f, err := fsys.Open(prefix + "test-user/test-repo/blob/test-branch/nested/file.txt")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.Content{}))
				c := f.(*ghfs.Content)
				Expect(c.Owner()).To(Equal("test-user"))
				Expect(c.Repository()).To(Equal("test-repo"))
				Expect(c.Branch()).To(Equal("test-branch"))
				Expect(c.Name()).To(Equal("nested/file.txt"))
				content, err := c.Content()
				Expect(err).NotTo(HaveOccurred())
				Expect(content.GetName()).To(Equal("file.txt"))
			})

			It("should return an error for a path with 3 segments", func() {
				_, err := fsys.Open(prefix + "test-user/test-repo/invalid")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ihfs.ErrNotExist))
			})
		},
	)
})

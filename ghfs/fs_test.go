package ghfs_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-github/v73/github"
	githubv82 "github.com/google/go-github/v82/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/ghfs"
)

// These test some invalid URLs. For example raw.githubusercontent.com/ does not
// follow the `owner/repo/tree/branch` pattern, and github.com/ does not follow
// the `owner/repo/blob/branch/path` pattern. I don't think that matters though,
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
						mock.GetReposReleasesTagsByOwnerByRepoByTag,
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

	Describe("Name", func() {
		It("should return 'github'", func() {
			fsys := ghfs.New()
			Expect(fsys.Name()).To(Equal("github"))
		})
	})

	Describe("Options", func() {
		It("should support WithAuthToken", func() {
			fsys := ghfs.New(ghfs.WithAuthToken("test-token"))
			Expect(fsys).NotTo(BeNil())
		})

		It("should support WithContextFunc", func() {
			called := false
			ctxFunc := func(f *ghfs.Fs, o ihfs.Operation) context.Context {
				called = true
				return context.Background()
			}

			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatch(
					mock.GetUsersByUsername,
					github.User{Name: github.Ptr("test-user")},
				),
			)
			DeferCleanup(s.Close)

			fsys := ghfs.New(
				ghfs.WithHttpClient(mockHttp),
				ghfs.WithContextFunc(ctxFunc),
			)

			_, _ = fsys.Open("test-user")
			Expect(called).To(BeTrue())
		})
	})

	Describe("API errors", func() {
		var fsys ihfs.FS

		BeforeEach(func() {
			// Mock server that returns 404 errors
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatchHandler(
					mock.GetUsersByUsername,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposBranchesByOwnerByRepoByBranch,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposContentsByOwnerByRepoByPath,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposReleasesByOwnerByRepoByReleaseId,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposReleasesAssetsByOwnerByRepoByAssetId,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			)
			DeferCleanup(s.Close)
			fsys = ghfs.New(ghfs.WithHttpClient(mockHttp))
		})

		It("should return error when openOwner fails", func() {
			_, err := fsys.Open("test-user")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when openRepository fails", func() {
			_, err := fsys.Open("test-user/test-repo")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when openBranch fails", func() {
			_, err := fsys.Open("test-user/test-repo/tree/test-branch")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when openContent fails", func() {
			_, err := fsys.Open("test-user/test-repo/blob/test-branch/file.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when openRelease fails", func() {
			_, err := fsys.Open("test-user/test-repo/releases/tag/v0.1.0")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when openAsset fails", func() {
			_, err := fsys.Open("test-user/test-repo/releases/tag/v0.1.0/asset.tar.gz")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("NewRequest errors", func() {
		It("should handle error when creating request with invalid base URL", func() {
			// Create a client with an invalid BaseURL to trigger NewRequest error
			client := githubv82.NewClient(nil)
			client.BaseURL.Path = "://invalid"
			fsys := ghfs.New(ghfs.WithClient(client))

			_, err := fsys.Open("test-user")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Empty path handling", func() {
		It("should return error for empty path after cleaning", func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatch(
					mock.GetUser,
					github.User{Name: github.Ptr("current-user")},
				),
			)
			DeferCleanup(s.Close)
			fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

			// Empty string should call openCurrent, not openOwner("")
			f, err := fsys.Open("")
			Expect(err).NotTo(HaveOccurred())
			Expect(f).To(BeAssignableToTypeOf(&ghfs.Owner{}))
		})

		It("should return error when openCurrent fails", func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatchHandler(
					mock.GetUser,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusUnauthorized)
						w.Write([]byte(`{"message": "Requires authentication"}`))
					}),
				),
			)
			DeferCleanup(s.Close)
			fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

			_, err := fsys.Open("")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Multiple accessor calls", func() {
		It("should support calling User() multiple times", func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatch(
					mock.GetUsersByUsername,
					github.User{Name: github.Ptr("test-user")},
				),
			)
			DeferCleanup(s.Close)
			fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

			f, err := fsys.Open("test-user")
			Expect(err).NotTo(HaveOccurred())
			owner := f.(*ghfs.Owner)

			// First call
			u1, err := owner.User()
			Expect(err).NotTo(HaveOccurred())
			Expect(u1.GetName()).To(Equal("test-user"))

			// Second call should work due to Seek(0, 0) in dec()
			u2, err := owner.User()
			Expect(err).NotTo(HaveOccurred())
			Expect(u2.GetName()).To(Equal("test-user"))
		})

		It("should support calling Repository() multiple times", func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatch(
					mock.GetReposByOwnerByRepo,
					github.Repository{
						Name: github.Ptr("test-repo"),
						Owner: &github.User{
							Name: github.Ptr("test-user"),
						},
					},
				),
			)
			DeferCleanup(s.Close)
			fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

			f, err := fsys.Open("test-user/test-repo")
			Expect(err).NotTo(HaveOccurred())
			repo := f.(*ghfs.Repository)

			// First call
			r1, err := repo.Repository()
			Expect(err).NotTo(HaveOccurred())
			Expect(r1.GetName()).To(Equal("test-repo"))

			// Second call should work
			r2, err := repo.Repository()
			Expect(err).NotTo(HaveOccurred())
			Expect(r2.GetName()).To(Equal("test-repo"))
		})
	})

	Describe("Invalid route patterns", func() {
		var fsys ihfs.FS

		BeforeEach(func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatchHandler(
					mock.GetReposReleasesTagsByOwnerByRepoByTag,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			)
			DeferCleanup(s.Close)
			fsys = ghfs.New(ghfs.WithHttpClient(mockHttp))
		})

		It("should return error for 4-segment path with invalid type", func() {
			// Not "tree" in position 2
			_, err := fsys.Open("owner/repo/invalid/branch")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ihfs.ErrNotExist))
		})

		It("should return error for 6-segment path with invalid releases type", func() {
			// Not "tag" or "download" in position 3 for releases
			_, err := fsys.Open("owner/repo/releases/invalid/tag/asset")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ihfs.ErrNotExist))
		})

		It("should return error for 6-segment path without tree or blob", func() {
			// Not "tree" or "blob" in position 2
			_, err := fsys.Open("owner/repo/invalid/branch/path/file")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ihfs.ErrNotExist))
		})
	})
})

package ghfs_test

import (
	"net/http"
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-github/v84/github"
	"github.com/unstoppablemango/go-github-mock/src/mock"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/ghfs"
)

// mockHTTPClient sets up a mock HTTP client with standard GitHub API responses.
func mockHTTPClient() (*http.Client, func()) {
	c, s := mock.NewMockedHTTPClientAndServer(
		mock.WithRequestMatch(
			mock.GetUsersByUsername,
			github.User{Name: github.Ptr("test-user")},
		),
		mock.WithRequestMatch(
			mock.GetReposByOwnerByRepo,
			github.Repository{
				Name:  github.Ptr("test-repo"),
				Owner: &github.User{Name: github.Ptr("test-user")},
			},
		),
		mock.WithRequestMatch(
			mock.GetReposReleasesTagsByOwnerByRepoByTag,
			github.RepositoryRelease{
				Name: github.Ptr("test-release"),
				Assets: []*github.ReleaseAsset{{
					ID:   github.Ptr(int64(1)),
					Name: github.Ptr("asset.tar.gz"),
				}},
			},
		),
		mock.WithRequestMatch(
			mock.GetReposReleasesAssetsByOwnerByRepoByAssetId,
			github.ReleaseAsset{Name: github.Ptr("asset.tar.gz")},
		),
		mock.WithRequestMatch(
			mock.GetReposBranchesByOwnerByRepoByBranch,
			github.Branch{Name: github.Ptr("test-branch")},
		),
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			github.RepositoryContent{Name: github.Ptr("file.txt")},
		),
	)
	return c, s.Close
}

var _ = Describe("Fs", func() {
	DescribeTableSubtree("Open web-style",
		Entry(nil, "https://github.com/"),
		Entry(nil, "github.com/"),
		func(prefix string) {
			var fsys ihfs.FS

			BeforeEach(func() {
				c, close := mockHTTPClient()
				DeferCleanup(close)
				fsys = ghfs.New(ghfs.WithHttpClient(c))
			})

			It("should open an owner path", func() {
				f, err := fsys.Open(prefix + "test-user")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open a repository path", func() {
				f, err := fsys.Open(prefix + "test-user/test-repo")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			DescribeTable("should open a release path",
				func(path string) {
					f, err := fsys.Open(prefix + path)

					Expect(err).NotTo(HaveOccurred())
					Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
				},
				Entry(nil, "test-user/test-repo/releases/tag/v0.1.0"),
				Entry(nil, "test-user/test-repo/releases/download/v0.1.0"),
			)

			DescribeTable("should open an asset path",
				func(path string) {
					f, err := fsys.Open(prefix + path)

					Expect(err).NotTo(HaveOccurred())
					Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
				},
				Entry(nil, "test-user/test-repo/releases/tag/v0.1.0/asset.tar.gz"),
				Entry(nil, "test-user/test-repo/releases/download/v0.1.0/asset.tar.gz"),
			)

			It("should open a branch path", func() {
				f, err := fsys.Open(prefix + "test-user/test-repo/tree/test-branch")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open a content path", func() {
				f, err := fsys.Open(prefix + "test-user/test-repo/blob/test-branch/file.txt")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open a nested content path", func() {
				f, err := fsys.Open(prefix + "test-user/test-repo/blob/test-branch/nested/file.txt")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			// It("should return an error for a path with 3 segments", func() {
			// 	_, err := fsys.Open(prefix + "test-user/test-repo/invalid")

			// 	Expect(err).To(HaveOccurred())
			// 	Expect(err).To(MatchError(ihfs.ErrNotExist))
			// })
		},
	)

	DescribeTableSubtree("Open raw-style",
		Entry(nil, "https://raw.githubusercontent.com/"),
		Entry(nil, "raw.githubusercontent.com/"),
		func(prefix string) {
			var fsys ihfs.FS

			BeforeEach(func() {
				c, close := mockHTTPClient()
				DeferCleanup(close)
				fsys = ghfs.New(ghfs.WithHttpClient(c))
			})

			It("should open an owner path", func() {
				f, err := fsys.Open(prefix + "test-user")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open a repository path", func() {
				f, err := fsys.Open(prefix + "test-user/test-repo")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open a branch path", func() {
				f, err := fsys.Open(prefix + "test-user/test-repo/test-branch")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open a content path", func() {
				f, err := fsys.Open(prefix + "test-user/test-repo/test-branch/file.txt")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open a nested content path", func() {
				f, err := fsys.Open(prefix + "test-user/test-repo/test-branch/nested/file.txt")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})
		},
	)

	DescribeTableSubtree("Open API pass-through",
		Entry(nil, "https://api.github.com/"),
		Entry(nil, "api.github.com/"),
		Entry("No prefix", ""),
		func(prefix string) {
			var fsys ihfs.FS

			BeforeEach(func() {
				c, close := mockHTTPClient()
				DeferCleanup(close)
				fsys = ghfs.New(ghfs.WithHttpClient(c))
			})

			It("should open an owner path", func() {
				f, err := fsys.Open(prefix + "users/test-user")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open a repository path", func() {
				f, err := fsys.Open(prefix + "repos/test-user/test-repo")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open a release path", func() {
				f, err := fsys.Open(prefix + "repos/test-user/test-repo/releases/tags/v0.1.0")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open an asset path", func() {
				f, err := fsys.Open(prefix + "repos/test-user/test-repo/releases/assets/1")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open a branch path", func() {
				f, err := fsys.Open(prefix + "repos/test-user/test-repo/branches/test-branch")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open a content path", func() {
				f, err := fsys.Open(prefix + "repos/test-user/test-repo/contents/file.txt?ref=test-branch")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})

			It("should open a nested content path", func() {
				f, err := fsys.Open(prefix + "repos/test-user/test-repo/contents/nested/file.txt?ref=test-branch")

				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(BeAssignableToTypeOf(&ghfs.File{}))
			})
		},
	)

	Describe("Name", func() {
		It("should return 'github'", func() {
			fsys := ghfs.New()
			Expect(fsys.Name()).To(Equal("github"))
		})
	})

	Describe("fstest", func() {
		It("should pass fstest.TestFS", func() {
			fsys := ghfs.New()
			Expect(fstest.TestFS(fsys)).To(Succeed())
		})
	})

	Describe("invalid paths", func() {
		DescribeTable("should return ErrInvalid",
			func(path string) {
				_, err := ghfs.New().Open(path)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ihfs.ErrInvalid))
			},
			Entry(nil, "/."),
			Entry(nil, "./."),
			Entry(nil, "/"),
		)

		It("should return error for unknown host", func() {
			_, err := ghfs.New().Open("https://gitlab.com/owner/repo")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("API errors", func() {
		var fsys ihfs.FS

		BeforeEach(func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatchHandler(
					mock.GetUsersByUsername,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposBranchesByOwnerByRepoByBranch,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposContentsByOwnerByRepoByPath,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposReleasesByOwnerByRepoByReleaseId,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposReleasesAssetsByOwnerByRepoByAssetId,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			)
			DeferCleanup(s.Close)
			fsys = ghfs.New(ghfs.WithHttpClient(mockHttp))
		})

		It("should return error when openOwner fails", func() {
			_, err := fsys.Open("users/test-user")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when openRepository fails", func() {
			_, err := fsys.Open("repos/test-user/test-repo")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when openBranch fails", func() {
			_, err := fsys.Open("repos/test-user/test-repo/branches/test-branch")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when openContent fails", func() {
			_, err := fsys.Open("repos/test-user/test-repo/contents/file.txt?ref=test-branch")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when openRelease fails", func() {
			_, err := fsys.Open("repos/test-user/test-repo/releases/tags/v0.1.0")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when openAsset fails", func() {
			_, err := fsys.Open("repos/test-user/test-repo/releases/assets/1")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Asset lookup by numeric release ID", func() {
		It("should resolve asset name via GetRelease", func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatch(
					mock.GetReposReleasesByOwnerByRepoByReleaseId,
					github.RepositoryRelease{
						Name: github.Ptr("test-release"),
						Assets: []*github.ReleaseAsset{
							{ID: github.Ptr(int64(1)), Name: github.Ptr("asset.tar.gz")},
						},
					},
				),
				mock.WithRequestMatch(
					mock.GetReposReleasesAssetsByOwnerByRepoByAssetId,
					github.ReleaseAsset{Name: github.Ptr("asset.tar.gz")},
				),
			)
			DeferCleanup(s.Close)
			fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

			f, err := fsys.Open("github.com/test-user/test-repo/releases/tag/12345/asset.tar.gz")
			Expect(err).NotTo(HaveOccurred())
			Expect(f).NotTo(BeNil())
		})
	})

	Describe("Asset lookup errors", func() {
		It("should return error when release lookup fails", func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatchHandler(
					mock.GetReposReleasesTagsByOwnerByRepoByTag,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			)
			DeferCleanup(s.Close)
			fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

			_, err := fsys.Open("github.com/test-user/test-repo/releases/tag/v0.1.0/asset.tar.gz")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when release decode fails", func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatchHandler(
					mock.GetReposReleasesTagsByOwnerByRepoByTag,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("invalid json"))
					}),
				),
			)
			DeferCleanup(s.Close)
			fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

			_, err := fsys.Open("github.com/test-user/test-repo/releases/tag/v0.1.0/asset.tar.gz")
			Expect(err).To(HaveOccurred())
		})

		It("should return ErrNotExist when asset name is not found in release", func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatch(
					mock.GetReposReleasesTagsByOwnerByRepoByTag,
					github.RepositoryRelease{
						Name:   github.Ptr("test-release"),
						Assets: []*github.ReleaseAsset{},
					},
				),
			)
			DeferCleanup(s.Close)
			fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

			_, err := fsys.Open("github.com/test-user/test-repo/releases/tag/v0.1.0/asset.tar.gz")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ihfs.ErrNotExist))
		})

		It("should return error when asset download fails", func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatch(
					mock.GetReposReleasesTagsByOwnerByRepoByTag,
					github.RepositoryRelease{
						Name: github.Ptr("test-release"),
						Assets: []*github.ReleaseAsset{
							{ID: github.Ptr(int64(1)), Name: github.Ptr("asset.tar.gz")},
						},
					},
				),
				mock.WithRequestMatchHandler(
					mock.GetReposReleasesAssetsByOwnerByRepoByAssetId,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			)
			DeferCleanup(s.Close)
			fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

			_, err := fsys.Open("github.com/test-user/test-repo/releases/tag/v0.1.0/asset.tar.gz")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("NewRequest errors", func() {
		It("should handle error when creating request with invalid base URL", func() {
			client := github.NewClient(nil)
			client.BaseURL.Path = "://invalid"
			fsys := ghfs.New(ghfs.WithClient(client))

			_, err := fsys.Open("users/test-user")
			Expect(err).To(HaveOccurred())
		})
	})
})

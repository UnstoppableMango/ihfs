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
			github.RepositoryRelease{Name: github.Ptr("test-release")},
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
	// Web-style prefixes: paths use owner/repo/tree/branch conventions.
	// github.com/ (schemeless) and https://github.com/ both route through fromWebURL.
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

			It("should return an error for a path with 3 segments", func() {
				_, err := fsys.Open(prefix + "test-user/test-repo/invalid")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ihfs.ErrNotExist))
			})
		},
	)

	// Raw-style prefixes: paths use owner/repo/branch/path conventions
	// (no tree/blob/releases keywords). 3 segments = branch, not an error.
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

	// API pass-through prefixes: paths must be valid GitHub API paths.
	// Query strings are preserved. No ErrNotExist for unknown paths — the
	// GitHub client returns a 404, not a path error.
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
				f, err := fsys.Open(prefix + "repos/test-user/test-repo/releases/assets/asset.tar.gz")

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

			_, _ = fsys.Open("users/test-user")
			Expect(called).To(BeTrue())
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
			_, err := fsys.Open("repos/test-user/test-repo/releases/assets/asset.tar.gz")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("NewRequest errors", func() {
		It("should handle error when creating request with invalid base URL", func() {
			client := githubv82.NewClient(nil)
			client.BaseURL.Path = "://invalid"
			fsys := ghfs.New(ghfs.WithClient(client))

			_, err := fsys.Open("users/test-user")
			Expect(err).To(HaveOccurred())
		})
	})
})

package ghfs_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-github/v84/github"
	"github.com/unstoppablemango/go-github-mock/src/mock"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/ghfs"
)

var _ = Describe("Open", func() {
	It("should return ErrNotImplemented for non-ghfs FS", func() {
		fsys := nonGhfsFS{}
		_, err := ghfs.Open(fsys, "users/test-user")
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ihfs.ErrNotImplemented))
	})
})

var _ = Describe("OpenOwner", func() {
	It("should return the user", func() {
		mockHttp, s := mock.NewMockedHTTPClientAndServer(
			mock.WithRequestMatch(
				mock.GetUsersByUsername,
				github.User{Name: github.Ptr("test-user")},
			),
		)
		DeferCleanup(s.Close)
		fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

		u, err := ghfs.OpenOwner(fsys, "test-user")
		Expect(err).NotTo(HaveOccurred())
		Expect(u.GetName()).To(Equal("test-user"))
	})
})

var _ = Describe("OpenRepository", func() {
	It("should return the repository", func() {
		mockHttp, s := mock.NewMockedHTTPClientAndServer(
			mock.WithRequestMatch(
				mock.GetReposByOwnerByRepo,
				github.Repository{Name: github.Ptr("test-repo")},
			),
		)
		DeferCleanup(s.Close)
		fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

		r, err := ghfs.OpenRepository(fsys, "test-user", "test-repo")
		Expect(err).NotTo(HaveOccurred())
		Expect(r.GetName()).To(Equal("test-repo"))
	})
})

var _ = Describe("OpenBranch", func() {
	It("should return the branch", func() {
		mockHttp, s := mock.NewMockedHTTPClientAndServer(
			mock.WithRequestMatch(
				mock.GetReposBranchesByOwnerByRepoByBranch,
				github.Branch{Name: github.Ptr("main")},
			),
		)
		DeferCleanup(s.Close)
		fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

		b, err := ghfs.OpenBranch(fsys, "test-user", "test-repo", "main")
		Expect(err).NotTo(HaveOccurred())
		Expect(b.GetName()).To(Equal("main"))
	})
})

var _ = Describe("OpenContent", func() {
	It("should return the content", func() {
		mockHttp, s := mock.NewMockedHTTPClientAndServer(
			mock.WithRequestMatch(
				mock.GetReposContentsByOwnerByRepoByPath,
				github.RepositoryContent{Name: github.Ptr("file.txt")},
			),
		)
		DeferCleanup(s.Close)
		fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

		c, err := ghfs.OpenContent(fsys, "test-user", "test-repo", "main", "file.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(c.GetName()).To(Equal("file.txt"))
	})
})

var _ = Describe("OpenRelease", func() {
	It("should return the release", func() {
		mockHttp, s := mock.NewMockedHTTPClientAndServer(
			mock.WithRequestMatch(
				mock.GetReposReleasesTagsByOwnerByRepoByTag,
				github.RepositoryRelease{Name: github.Ptr("v1.0.0")},
			),
		)
		DeferCleanup(s.Close)
		fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

		r, err := ghfs.OpenRelease(fsys, "test-user", "test-repo", "v1.0.0")
		Expect(err).NotTo(HaveOccurred())
		Expect(r.GetName()).To(Equal("v1.0.0"))
	})
})

var _ = Describe("OpenOwner errors", func() {
	It("should return error when Open fails", func() {
		mockHttp, s := mock.NewMockedHTTPClientAndServer(
			mock.WithRequestMatchHandler(
				mock.GetUsersByUsername,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"message": "Not Found"}`))
				}),
			),
		)
		DeferCleanup(s.Close)
		fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

		_, err := ghfs.OpenOwner(fsys, "test-user")
		Expect(err).To(HaveOccurred())
	})

	It("should return error when decode fails", func() {
		mockHttp, s := mock.NewMockedHTTPClientAndServer(
			mock.WithRequestMatchHandler(
				mock.GetUsersByUsername,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("invalid json"))
				}),
			),
		)
		DeferCleanup(s.Close)
		fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

		_, err := ghfs.OpenOwner(fsys, "test-user")
		Expect(err).To(HaveOccurred())
	})
})

// nonGhfsFS is a minimal ihfs.FS that is not a *ghfs.Fs.
type nonGhfsFS struct{}

func (nonGhfsFS) Open(name string) (ihfs.File, error) {
	return nil, fmt.Errorf("not implemented")
}

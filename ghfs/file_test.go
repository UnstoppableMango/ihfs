package ghfs_test

import (
	"bytes"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-github/v73/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/ghfs"
)

var _ = Describe("File", func() {
	Describe("Close", func() {
		It("should return nil", func() {
			owner := &ghfs.Owner{}
			err := owner.Close()
			Expect(err).To(BeNil())

			repo := &ghfs.Repository{}
			err = repo.Close()
			Expect(err).To(BeNil())

			content := &ghfs.Content{}
			err = content.Close()
			Expect(err).To(BeNil())

			asset := &ghfs.Asset{}
			err = asset.Close()
			Expect(err).To(BeNil())

			release := &ghfs.Release{}
			err = release.Close()
			Expect(err).To(BeNil())

			branch := &ghfs.Branch{}
			err = branch.Close()
			Expect(err).To(BeNil())
		})
	})

	Describe("File methods", func() {
		var file ihfs.File

		BeforeEach(func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatch(
					mock.GetUsersByUsername,
					github.User{Name: github.Ptr("test-user")},
				),
			)
			DeferCleanup(s.Close)

			fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))
			var err error
			file, err = fsys.Open("test-user")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should support Close", func() {
			err := file.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should support Stat", func() {
			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Name()).To(Equal("test-user"))
		})

		It("should return size", func() {
			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Size()).To(BeNumerically(">", 0))
		})

		It("should return sys", func() {
			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Sys()).NotTo(BeNil())
			Expect(info.Sys()).To(BeAssignableToTypeOf(&bytes.Reader{}))
		})

		It("should panic on IsDir", func() {
			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { info.IsDir() }).To(Panic())
		})

		It("should panic on ModTime", func() {
			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { info.ModTime() }).To(Panic())
		})

		It("should panic on Mode", func() {
			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(func() { info.Mode() }).To(Panic())
		})
	})

	Describe("File type decode errors", func() {
		var fsys ihfs.FS

		BeforeEach(func() {
			// Mock server that returns invalid JSON
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatchHandler(
					mock.GetUsersByUsername,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("invalid json"))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("invalid json"))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposReleasesTagsByOwnerByRepoByTag,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("invalid json"))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposReleasesAssetsByOwnerByRepoByAssetId,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("invalid json"))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposBranchesByOwnerByRepoByBranch,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("invalid json"))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposContentsByOwnerByRepoByPath,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("invalid json"))
					}),
				),
			)
			DeferCleanup(s.Close)
			fsys = ghfs.New(ghfs.WithHttpClient(mockHttp))
		})

		It("should return error when User decode fails", func() {
			f, err := fsys.Open("test-user")
			Expect(err).NotTo(HaveOccurred())
			owner := f.(*ghfs.Owner)
			_, err = owner.User()
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Repository decode fails", func() {
			f, err := fsys.Open("test-user/test-repo")
			Expect(err).NotTo(HaveOccurred())
			repo := f.(*ghfs.Repository)
			_, err = repo.Repository()
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Release decode fails", func() {
			f, err := fsys.Open("test-user/test-repo/releases/tag/v0.1.0")
			Expect(err).NotTo(HaveOccurred())
			release := f.(*ghfs.Release)
			_, err = release.Release()
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Asset decode fails", func() {
			f, err := fsys.Open("test-user/test-repo/releases/tag/v0.1.0/asset.tar.gz")
			Expect(err).NotTo(HaveOccurred())
			asset := f.(*ghfs.Asset)
			_, err = asset.Asset()
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Branch decode fails", func() {
			f, err := fsys.Open("test-user/test-repo/tree/test-branch")
			Expect(err).NotTo(HaveOccurred())
			branch := f.(*ghfs.Branch)
			_, err = branch.Branch()
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Content decode fails", func() {
			f, err := fsys.Open("test-user/test-repo/blob/test-branch/file.txt")
			Expect(err).NotTo(HaveOccurred())
			content := f.(*ghfs.Content)
			_, err = content.Content()
			Expect(err).To(HaveOccurred())
		})
	})
})

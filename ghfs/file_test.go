package ghfs_test

import (
	"bytes"
	"io/fs"
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
			f := &ghfs.File{}
			err := f.Close()
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
			file, err = fsys.Open("users/test-user")
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

		It("should return false for IsDir", func() {
			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeFalse())
		})

		It("should return zero time for ModTime", func() {
			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.ModTime().IsZero()).To(BeTrue())
		})

		It("should return read-only mode for Mode", func() {
			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(fs.FileMode(0444)))
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
			f, err := ghfs.Open(fsys, "users/test-user")
			Expect(err).NotTo(HaveOccurred())

			var owner github.User
			err = f.Decode(&owner)
			Expect(err).To(HaveOccurred())
		})
	})
})

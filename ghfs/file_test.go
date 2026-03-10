package ghfs_test

import (
	"io"
	"io/fs"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-github/v84/github"
	"github.com/unstoppablemango/go-github-mock/src/mock"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/ghfs"
)

var _ = Describe("File", func() {
	Describe("Close", func() {
		It("should return nil", func() {
			f := &ghfs.File{}
			err := f.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Read", func() {
		It("should return 0 when rc is nil", func() {
			f := &ghfs.File{}
			n, err := f.Read(make([]byte, 10))
			Expect(n).To(Equal(0))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("ReadDir", func() {
		It("should return error for non-dir file", func() {
			f := &ghfs.File{}
			entries, err := f.ReadDir(1)
			Expect(entries).To(BeEmpty())
			Expect(err).To(MatchError(fs.ErrInvalid))
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

		It("should return -1 for size", func() {
			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Size()).To(Equal(int64(-1)))
		})

		It("should return sys as io.ReadCloser", func() {
			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			sys := info.Sys()
			Expect(sys).NotTo(BeNil())
			_, ok := sys.(io.ReadCloser)
			Expect(ok).To(BeTrue())
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

	Describe("Name with query string", func() {
		It("should strip query string from name", func() {
			mockHttp, s := mock.NewMockedHTTPClientAndServer(
				mock.WithRequestMatch(
					mock.GetReposContentsByOwnerByRepoByPath,
					github.RepositoryContent{Name: github.Ptr("file.txt")},
				),
			)
			DeferCleanup(s.Close)
			fsys := ghfs.New(ghfs.WithHttpClient(mockHttp))

			f, err := fsys.Open("repos/test-user/test-repo/contents/file.txt?ref=main")
			Expect(err).NotTo(HaveOccurred())

			info, err := f.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Name()).To(Equal("file.txt"))
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
						_, _ = w.Write([]byte("invalid json"))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("invalid json"))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposReleasesTagsByOwnerByRepoByTag,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("invalid json"))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposReleasesAssetsByOwnerByRepoByAssetId,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("invalid json"))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposBranchesByOwnerByRepoByBranch,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("invalid json"))
					}),
				),
				mock.WithRequestMatchHandler(
					mock.GetReposContentsByOwnerByRepoByPath,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("invalid json"))
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

	Describe("Dir", func() {
		var dir *ghfs.File

		BeforeEach(func() {
			f, err := ghfs.New().Open(".")
			Expect(err).NotTo(HaveOccurred())
			var ok bool
			dir, ok = f.(*ghfs.File)
			Expect(ok).To(BeTrue())
		})

		It("should return error on Read", func() {
			n, err := dir.Read(make([]byte, 10))
			Expect(n).To(Equal(0))
			Expect(err).To(MatchError(fs.ErrInvalid))
		})

		It("should succeed on Close", func() {
			Expect(dir.Close()).To(Succeed())
		})

		It("should return FileInfo from Stat", func() {
			info, err := dir.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
		})

		It("should return io.EOF on ReadDir with n>0", func() {
			entries, err := dir.ReadDir(1)
			Expect(entries).To(BeEmpty())
			Expect(err).To(Equal(io.EOF))
		})

		It("should return empty slice on ReadDir with n<=0", func() {
			entries, err := dir.ReadDir(-1)
			Expect(entries).To(BeEmpty())
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("FileInfo", func() {
			var info fs.FileInfo

			BeforeEach(func() {
				var err error
				info, err = dir.Stat()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return '.' for Name", func() {
				Expect(info.Name()).To(Equal("."))
			})

			It("should return true for IsDir", func() {
				Expect(info.IsDir()).To(BeTrue())
			})

			It("should return a directory mode", func() {
				Expect(info.Mode()).To(Equal(fs.ModeDir | 0555))
			})

			It("should return zero time for ModTime", func() {
				Expect(info.ModTime()).To(Equal(time.Time{}))
			})

			It("should return 0 for Size", func() {
				Expect(info.Size()).To(Equal(int64(0)))
			})

			It("should return nil for Sys", func() {
				Expect(info.Sys()).To(BeNil())
			})
		})
	})
})

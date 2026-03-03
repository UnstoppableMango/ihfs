package ghfs_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-github/v84/github"
	"github.com/unstoppablemango/go-github-mock/src/mock"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/ghfs"
)

var _ = Describe("Options", func() {
	It("should support WithAuthToken", func() {
		fsys := ghfs.New(ghfs.WithAuthToken("test-token"))
		Expect(fsys).NotTo(BeNil())
	})

	It("should apply auth token when WithHttpClient follows WithAuthToken", func() {
		var capturedHeader string
		mockHttp, s := mock.NewMockedHTTPClientAndServer(
			mock.WithRequestMatchHandler(
				mock.GetUsersByUsername,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					capturedHeader = r.Header.Get("Authorization")
					_, _ = w.Write([]byte(`{}`))
				}),
			),
		)
		DeferCleanup(s.Close)

		fsys := ghfs.New(
			ghfs.WithAuthToken("test-token"),
			ghfs.WithHttpClient(mockHttp),
		)

		_, _ = fsys.Open("users/test-user")
		Expect(capturedHeader).To(ContainSubstring("test-token"))
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

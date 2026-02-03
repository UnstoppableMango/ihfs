package ghfs

import (
	"net/http"

	"github.com/google/go-github/v82/github"
)

type Option func(*Fs)

func WithClient(client *github.Client) Option {
	return func(f *Fs) {
		f.client = client
	}
}

func WithHttpClient(client *http.Client) Option {
	return WithClient(github.NewClient(client))
}

func WithContextFunc(fn ContextFunc) Option {
	return func(f *Fs) {
		f.ctxFn = fn
	}
}

func WithAuthToken(token string) Option {
	return func(f *Fs) {
		if f.client == nil {
			f.client = github.NewClient(nil)
		}
		f.setAuthToken(token)
	}
}

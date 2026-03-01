package corfs

import (
	"time"

	"github.com/unstoppablemango/ihfs/union"
)

// Option configures a corfs [Fs].
type Option func(*Fs)

// WithCacheTime sets the cache duration for the corfs [Fs].
func WithCacheTime(cacheTime time.Duration) Option {
	return func(f *Fs) {
		f.cacheTime = cacheTime
	}
}

// WithMergeStrategy sets the merge strategy for the corfs [Fs].
func WithMergeStrategy(strategy union.MergeStrategy) Option {
	return func(f *Fs) {
		f.fopts = append(f.fopts,
			union.WithMergeStrategy(strategy),
		)
	}
}

// WithDefaultMergeStrategy sets the default merge strategy for the corfs [Fs].
func WithDefaultMergeStrategy() Option {
	return WithMergeStrategy(union.DefaultMergeStrategy)
}

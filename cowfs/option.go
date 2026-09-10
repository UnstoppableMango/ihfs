package cowfs

import "github.com/unstoppablemango/ihfs/union"

// Option configures a cowfs [Fs].
type Option func(*Fs)

// WithMergeStrategy sets the merge strategy for the cowfs [Fs].
func WithMergeStrategy(strategy union.MergeStrategy) Option {
	return func(f *Fs) {
		f.fopts = append(f.fopts,
			union.WithMergeStrategy(strategy),
		)
	}
}

// WithDefaultMergeStrategy sets the default merge strategy for the cowfs [Fs].
func WithDefaultMergeStrategy() Option {
	return WithMergeStrategy(union.DefaultMergeStrategy)
}

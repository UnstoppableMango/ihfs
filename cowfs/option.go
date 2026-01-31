package cowfs

import "github.com/unstoppablemango/ihfs/union"

type Option func(*Fs)

func WithMergeStrategy(strategy union.MergeStrategy) Option {
	return func(f *Fs) {
		f.fopts = append(f.fopts,
			union.WithMergeStrategy(strategy),
		)
	}
}

func WithDefaultMergeStrategy() Option {
	return WithMergeStrategy(union.DefaultMergeStrategy)
}

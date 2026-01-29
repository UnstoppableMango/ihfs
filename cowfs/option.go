package cowfs

type Option func(*Fs)

func WithMergeStrategy(strategy MergeStrategy) Option {
	return func(f *Fs) {
		f.merge = strategy
	}
}

func WithDefaultMergeStrategy() Option {
	return WithMergeStrategy(DefaultMergeStrategy)
}

package union

type Option func(*File)

func WithMergeStrategy(merge MergeStrategy) Option {
	return func(f *File) {
		f.merge = merge
	}
}

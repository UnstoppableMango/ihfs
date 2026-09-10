package union

// Option configures a union [File].
type Option func(*File)

// WithMergeStrategy sets the merge strategy for a union file's directory reading.
func WithMergeStrategy(merge MergeStrategy) Option {
	return func(f *File) {
		f.merge = merge
	}
}

package ignore

import (
	"iter"
	"slices"
)

type File []Pattern

func (f File) Patterns() []Pattern {
	return f
}

func (f File) Iter() iter.Seq[Pattern] {
	return slices.Values(f)
}

// Ignored returns true if filePath should be blocked by the patterns.
// Patterns are evaluated in order; a negation pattern overrides prior matches.
func (f File) Ignored(filePath string) bool {
	result := false
	for pattern := range f.Iter() {
		if m := pattern.Match(filePath); m != nil {
			result = *m
		}
	}
	return result
}

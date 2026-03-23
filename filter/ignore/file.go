package ignore

import (
	"slices"
)

type File []Pattern

func (f File) Patterns() []Pattern {
	return f
}

// Ignored returns true if filePath should be blocked by the patterns.
// Patterns are evaluated in order; a negation pattern overrides prior matches.
func (f File) Ignored(filePath string) bool {
	result := false
	for pattern := range slices.Values(f) {
		if pattern.Match(filePath) {
			result = !pattern.negated
		}
	}
	return result
}

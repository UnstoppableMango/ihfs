package filter

import (
	"io/fs"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/filter/ignore"
	"github.com/unstoppablemango/ihfs/op"
)

// Ignore creates an [ihfs.FilterFunc] that blocks operations on files matched
// by any of the gitignore-style patterns in lines.
// Blank lines and lines starting with '#' are treated as comments and ignored.
// A leading '!' negates a pattern, re-including a previously ignored file.
// A trailing '/' restricts the pattern to directories only; since directories
// always pass through to allow traversal, such patterns have no effect here.
// Patterns containing '/' are anchored to the filesystem root; patterns without
// '/' match against the file's base name at any depth.
// The wildcards '*', '?', '[...]', and '**' follow gitignore semantics.
func Ignore(lines []string) ihfs.FilterFunc {
	var patterns []ignore.Pattern
	for _, line := range lines {
		if p := ignore.Parse(line); p != nil {
			patterns = append(patterns, *p)
		}
	}

	return func(f *ihfs.FilterFS, o op.Operation) error {
		name, ok := op.Name(o)
		if !ok {
			return nil
		}
		if info, err := fs.Stat(f.Base(), name); err == nil && info.IsDir() {
			return nil
		}
		if ignore.Ignored(patterns, name) {
			return ihfs.ErrPermission
		}
		return nil
	}
}

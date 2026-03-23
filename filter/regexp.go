package filter

import (
	"io/fs"
	"regexp"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/op"
)

// NameRegex creates an [ihfs.FilterFunc] that allows operations on files whose
// name matches re, returning [ihfs.ErrPermission] otherwise.
// Operations that target a directory always pass through, as do operations
// without a Name field (e.g. [op.Glob]).
func NameRegex(re *regexp.Regexp) ihfs.FilterFunc {
	return func(f *ihfs.FilterFS, o op.Operation) error {
		name, ok := op.Name(o)
		if !ok {
			return nil
		}
		if re.MatchString(name) {
			return nil
		}

		// Directories always pass through (afero parity)
		if info, err := fs.Stat(f.Base(), name); err == nil && info.IsDir() {
			return nil
		}

		return ihfs.ErrPermission
	}
}

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
	return func(f *ihfs.FilterFS, o ihfs.Operation) error {
		var name string
		switch v := o.(type) {
		case op.Open:
			name = v.Name
		case op.Stat:
			name = v.Name
		case op.ReadDir:
			name = v.Name
		case op.Lstat:
			name = v.Name
		case op.ReadFile:
			name = v.Name
		case op.ReadLink:
			name = v.Name
		case op.WriteFile:
			name = v.Name
		case op.Remove:
			name = v.Name
		case op.RemoveAll:
			name = v.Name
		default:
			return nil // op.Glob and unknown ops pass through
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

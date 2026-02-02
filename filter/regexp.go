package filter

import (
	"regexp"

	"github.com/unstoppablemango/ihfs"
)

func NameRegex(re *regexp.Regexp) ihfs.FilterFunc {
	return func(_ *ihfs.FilterFS, o ihfs.Operation) error {
		return nil
	}
}

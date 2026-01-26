package try

import (
	"fmt"

	"github.com/unstoppablemango/ihfs"
)

// Seek attempts to call Seek on the given File.
// If the File does not implement [ihfs.Seeker], Seek returns [ErrNotSupported].
func Seek(f ihfs.File, offset int64, whence int) (int64, error) {
	if seeker, ok := f.(ihfs.Seeker); ok {
		return seeker.Seek(offset, whence)
	}
	return 0, fmt.Errorf("seek: %w", ErrNotSupported)
}

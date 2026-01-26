package try

import (
	"fmt"

	"github.com/unstoppablemango/ihfs"
)

// Seek attempts to call Seek on the given File.
// If the File does not implement [ihfs.Seeker], Seek returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Seek(f ihfs.File, offset int64, whence int) (int64, error) {
	if seeker, ok := f.(ihfs.Seeker); ok {
		return seeker.Seek(offset, whence)
	}
	return 0, fmt.Errorf("seek: %w", ErrNotSupported)
}

// Write attempts to call Write on the given File.
// If the File does not implement [ihfs.Writer], Write returns
// an error that can be checked with [errors.Is] for [ErrNotSupported].
func Write(f ihfs.File, p []byte) (int, error) {
	if writer, ok := f.(ihfs.Writer); ok {
		return writer.Write(p)
	}
	return 0, fmt.Errorf("write: %w", ErrNotSupported)
}

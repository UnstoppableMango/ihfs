// Package osfs provides an OS-backed filesystem implementation.
package osfs

import (
	"github.com/unmango/go/os"
	"github.com/unstoppablemango/ihfs"
)

// Default is the default OS filesystem backed by [os.System].
var Default ihfs.OsFS = Fs{os.System}

// Fs is an OS-backed filesystem.
type Fs struct{ os.Fs }

// New creates a new OS filesystem backed by [os.System].
func New() ihfs.OsFS {
	return Fs{os.System}
}

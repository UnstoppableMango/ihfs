package osfs

import (
	"github.com/unmango/go/os"
	"github.com/unstoppablemango/ihfs"
)

var Default ihfs.OsFS = Fs{os.System}

type Fs struct{ os.Fs }

func New() ihfs.OsFS {
	return Fs{os.System}
}

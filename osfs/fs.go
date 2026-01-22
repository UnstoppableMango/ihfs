package osfs

import (
	"github.com/unmango/go/os"
	"github.com/unstoppablemango/ihfs"
)

var Default ihfs.Os = Fs{os.System}

type Fs struct{ os.Fs }

func New() ihfs.Os {
	return Fs{os.System}
}

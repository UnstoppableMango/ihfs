package osfs

import (
	"github.com/unmango/go/os"
	"github.com/unstoppablemango/ihfs"
)

type Fs struct{ os.Fs }

func New() ihfs.Os {
	return &Fs{os.System}
}

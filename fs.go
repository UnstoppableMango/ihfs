package ihfs

import (
	"io/fs"

	"github.com/unmango/go/os"
)

type (
	FS       = fs.FS
	Glob     = fs.GlobFS
	ReadDir  = fs.ReadDirFS
	ReadFile = fs.ReadFileFS
	ReadLink = fs.ReadLinkFS
	Stat     = fs.StatFS
	Sub      = fs.SubFS

	FileInfo = fs.FileInfo
	DirEntry = fs.DirEntry
)

type Os interface {
	os.Fs
}

var _ FS = (Os)(nil)

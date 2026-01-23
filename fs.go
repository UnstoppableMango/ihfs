package ihfs

import (
	"io/fs"
	"time"

	"github.com/unmango/go/os"
)

type (
	FS       = fs.FS
	Glob     = fs.GlobFS
	Os       = os.Fs
	ReadDir  = fs.ReadDirFS
	ReadFile = fs.ReadFileFS
	ReadLink = fs.ReadLinkFS
	Stat     = fs.StatFS
	Sub      = fs.SubFS

	DirEntry = fs.DirEntry
	File     = fs.File
	FileInfo = fs.FileInfo
	FileMode = fs.FileMode
)

// Ensure interface compliance.
var _ FS = (Os)(nil)

type Copy interface {
	FS
	Copy(name string, dest FS) error
}

type Mkdir interface {
	FS
	Mkdir(name string, perm FileMode) error
}

type Chmod interface {
	FS
	Chmod(name string, mode FileMode) error
}

type Chown interface {
	FS
	Chown(name string, uid, gid int) error
}

type Chtimes interface {
	FS
	Chtimes(name string, atime, mtime time.Time) error
}

// ReadDirNames reads the named directory and returns a list of names sorted by filename.
func ReadDirNames(f FS, name string) ([]string, error) {
	entries, err := dirEntries(f, name)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names, nil
}

func dirEntries(f FS, name string) ([]DirEntry, error) {
	if dirfs, ok := f.(ReadDir); ok {
		return dirfs.ReadDir(name)
	}
	return fs.ReadDir(f, name)
}

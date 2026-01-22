package ihfs

import (
	"io/fs"

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

	File     = fs.File
	FileInfo = fs.FileInfo
	DirEntry = fs.DirEntry
)

// Ensure interface compliance.
var _ FS = (Os)(nil)

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

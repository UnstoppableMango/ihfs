package protofsv1alpha1

import (
	"io/fs"

	filev1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/dev/unmango/file/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// dirEntry wraps a proto FileInfo message and implements fs.DirEntry.
type dirEntry struct {
	proto *filev1alpha1.FileInfo
}

// IsDir implements fs.DirEntry.
func (d dirEntry) IsDir() bool {
	return d.proto.IsDir
}

// Name implements fs.DirEntry.
func (d dirEntry) Name() string {
	return d.proto.Name
}

// Type implements fs.DirEntry.
func (d dirEntry) Type() fs.FileMode {
	return fs.FileMode(d.proto.Mode) & fs.ModeType
}

// Info implements fs.DirEntry.
func (d dirEntry) Info() (fs.FileInfo, error) {
	return fromProtoFileInfo(d.proto), nil
}

// DirEntryToFileInfo converts an fs.DirEntry to a proto FileInfo.
func DirEntryToFileInfo(entry fs.DirEntry) *filev1alpha1.FileInfo {
	return dirEntryToFileInfo(entry)
}

// FileInfosToDirEntries converts a slice of proto FileInfo messages to []fs.DirEntry.
func FileInfosToDirEntries(infos []*filev1alpha1.FileInfo) []fs.DirEntry {
	return fileInfosToDirEntries(infos)
}

func dirEntryToFileInfo(entry fs.DirEntry) *filev1alpha1.FileInfo {
	fi := &filev1alpha1.FileInfo{
		Name:  entry.Name(),
		IsDir: entry.IsDir(),
		Mode:  filev1alpha1.FileMode(entry.Type()),
	}

	if info, err := entry.Info(); err == nil && info != nil {
		fi.Size = info.Size()
		fi.Mode = filev1alpha1.FileMode(info.Mode())
		fi.ModTime = timestamppb.New(info.ModTime())
		fi.IsDir = info.IsDir()
	}

	return fi
}

func fileInfosToDirEntries(infos []*filev1alpha1.FileInfo) []fs.DirEntry {
	result := make([]fs.DirEntry, len(infos))
	for i, info := range infos {
		result[i] = dirEntry{proto: info}
	}

	return result
}

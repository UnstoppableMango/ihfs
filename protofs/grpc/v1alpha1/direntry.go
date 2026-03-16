package protofsv1alpha1

import (
	"io/fs"

	ihfsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/ihfs/v1alpha1"
)

// dirEntry wraps a proto DirEntry message and implements fs.DirEntry.
type dirEntry struct {
	proto *ihfsv1alpha1.DirEntry
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
	return fs.FileMode(d.proto.Type)
}

// Info implements fs.DirEntry.
func (d dirEntry) Info() (fs.FileInfo, error) {
	if d.proto.Info == nil {
		return nil, nil
	}

	return fromProtoFileInfo(d.proto.Info), nil
}

// ToProtoDirEntry converts an fs.DirEntry to its proto representation.
func ToProtoDirEntry(entry fs.DirEntry) *ihfsv1alpha1.DirEntry {
	return toProtoDirEntry(entry)
}

// FromProtoDirEntries converts a slice of proto DirEntry messages to []fs.DirEntry.
func FromProtoDirEntries(entries []*ihfsv1alpha1.DirEntry) []fs.DirEntry {
	return fromProtoDirEntries(entries)
}

func toProtoDirEntry(entry fs.DirEntry) *ihfsv1alpha1.DirEntry {
	proto := &ihfsv1alpha1.DirEntry{
		Name:  entry.Name(),
		IsDir: entry.IsDir(),
		Type:  uint32(entry.Type()),
	}

	if info, err := entry.Info(); err == nil && info != nil {
		proto.Info = toProtoFileInfo(info)
	}

	return proto
}

func fromProtoDirEntries(entries []*ihfsv1alpha1.DirEntry) []fs.DirEntry {
	result := make([]fs.DirEntry, len(entries))
	for i, e := range entries {
		result[i] = dirEntry{proto: e}
	}

	return result
}

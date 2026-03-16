package protofsv1alpha1

import (
	"io/fs"
	"time"

	ihfsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/ihfs/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// fileInfo wraps a proto FileInfo message and implements fs.FileInfo.
type fileInfo struct {
	proto *ihfsv1alpha1.FileInfo
}

// IsDir implements fs.FileInfo.
func (f fileInfo) IsDir() bool {
	return f.proto.IsDir
}

// ModTime implements fs.FileInfo.
func (f fileInfo) ModTime() time.Time {
	return f.proto.ModTime.AsTime()
}

// Mode implements fs.FileInfo.
func (f fileInfo) Mode() fs.FileMode {
	return fs.FileMode(f.proto.Mode)
}

// Name implements fs.FileInfo.
func (f fileInfo) Name() string {
	return f.proto.Name
}

// Size implements fs.FileInfo.
func (f fileInfo) Size() int64 {
	return f.proto.Size
}

// Sys implements fs.FileInfo.
func (f fileInfo) Sys() any {
	return nil
}

// ToProtoFileInfo converts an fs.FileInfo to its proto representation.
func ToProtoFileInfo(fi fs.FileInfo) *ihfsv1alpha1.FileInfo {
	return toProtoFileInfo(fi)
}

// FromProtoFileInfo converts a proto FileInfo to an fs.FileInfo.
func FromProtoFileInfo(fi *ihfsv1alpha1.FileInfo) fs.FileInfo {
	return fromProtoFileInfo(fi)
}

func toProtoFileInfo(fi fs.FileInfo) *ihfsv1alpha1.FileInfo {
	return &ihfsv1alpha1.FileInfo{
		Name:    fi.Name(),
		Size:    fi.Size(),
		Mode:    uint32(fi.Mode()),
		ModTime: timestamppb.New(fi.ModTime()),
		IsDir:   fi.IsDir(),
	}
}

func fromProtoFileInfo(fi *ihfsv1alpha1.FileInfo) fs.FileInfo {
	return fileInfo{proto: fi}
}

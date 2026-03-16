package protofsv1alpha1_test

import (
	"io/fs"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ihfsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/ihfs/v1alpha1"
	protofsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/grpc/v1alpha1"
	"github.com/unstoppablemango/ihfs/testfs"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ = Describe("FileInfo", func() {
	Describe("ToProtoFileInfo", func() {
		It("should convert a FileInfo to proto", func() {
			now := time.Now().UTC().Truncate(time.Second)
			fi := testfs.NewFileInfo("test.txt")
			fi.SizeFunc = func() int64 { return 42 }
			fi.ModeFunc = func() fs.FileMode { return 0o644 }
			fi.ModTimeFunc = func() time.Time { return now }
			fi.IsDirFunc = func() bool { return false }

			proto := protofsv1alpha1.ToProtoFileInfo(fi)

			Expect(proto.Name).To(Equal("test.txt"))
			Expect(proto.Size).To(Equal(int64(42)))
			Expect(proto.Mode).To(Equal(uint32(0o644)))
			Expect(proto.ModTime.AsTime()).To(BeTemporally("~", now, time.Second))
			Expect(proto.IsDir).To(BeFalse())
		})

		It("should convert a directory FileInfo to proto", func() {
			fi := testfs.NewFileInfo("subdir")
			fi.IsDirFunc = func() bool { return true }

			proto := protofsv1alpha1.ToProtoFileInfo(fi)

			Expect(proto.Name).To(Equal("subdir"))
			Expect(proto.IsDir).To(BeTrue())
		})
	})

	Describe("FromProtoFileInfo", func() {
		It("should convert a proto FileInfo to FileInfo", func() {
			now := time.Now().UTC().Truncate(time.Second)
			proto := &ihfsv1alpha1.FileInfo{
				Name:    "test.txt",
				Size:    42,
				Mode:    uint32(0o644),
				ModTime: timestamppb.New(now),
				IsDir:   false,
			}

			fi := protofsv1alpha1.FromProtoFileInfo(proto)

			Expect(fi.Name()).To(Equal("test.txt"))
			Expect(fi.Size()).To(Equal(int64(42)))
			Expect(fi.Mode()).To(Equal(fs.FileMode(0o644)))
			Expect(fi.ModTime()).To(BeTemporally("~", now, time.Second))
			Expect(fi.IsDir()).To(BeFalse())
			Expect(fi.Sys()).To(BeNil())
		})

		It("should convert a directory proto FileInfo", func() {
			proto := &ihfsv1alpha1.FileInfo{
				Name:  "subdir",
				IsDir: true,
			}

			fi := protofsv1alpha1.FromProtoFileInfo(proto)

			Expect(fi.Name()).To(Equal("subdir"))
			Expect(fi.IsDir()).To(BeTrue())
		})
	})
})

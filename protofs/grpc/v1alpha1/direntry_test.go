package protofsv1alpha1_test

import (
	"errors"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ihfsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/ihfs/v1alpha1"
	protofsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/grpc/v1alpha1"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("DirEntry", func() {
	Describe("ToProtoDirEntry", func() {
		It("should convert a DirEntry to proto with info", func() {
			entry := testfs.NewDirEntry("subdir", true)

			proto := protofsv1alpha1.ToProtoDirEntry(entry)

			Expect(proto.Name).To(Equal("subdir"))
			Expect(proto.IsDir).To(BeTrue())
			Expect(proto.Info).NotTo(BeNil())
		})

		It("should convert a file DirEntry to proto", func() {
			entry := testfs.NewDirEntry("file.txt", false)

			proto := protofsv1alpha1.ToProtoDirEntry(entry)

			Expect(proto.Name).To(Equal("file.txt"))
			Expect(proto.IsDir).To(BeFalse())
		})

		It("should handle DirEntry with failing Info()", func() {
			entry := &errInfoDirEntry{name: "broken"}

			proto := protofsv1alpha1.ToProtoDirEntry(entry)

			Expect(proto.Name).To(Equal("broken"))
			Expect(proto.Info).To(BeNil())
		})
	})

	Describe("FromProtoDirEntries", func() {
		It("should convert proto DirEntries to DirEntries", func() {
			protos := []*ihfsv1alpha1.DirEntry{
				{Name: "dir1", IsDir: true, Type: uint32(fs.ModeDir)},
				{Name: "file.txt", IsDir: false},
			}

			entries := protofsv1alpha1.FromProtoDirEntries(protos)

			Expect(entries).To(HaveLen(2))
			Expect(entries[0].Name()).To(Equal("dir1"))
			Expect(entries[0].IsDir()).To(BeTrue())
			Expect(entries[0].Type()).To(Equal(fs.ModeDir))
			Expect(entries[1].Name()).To(Equal("file.txt"))
			Expect(entries[1].IsDir()).To(BeFalse())
		})

		It("should return Info from proto DirEntry with info", func() {
			protos := []*ihfsv1alpha1.DirEntry{
				{
					Name:  "test.txt",
					IsDir: false,
					Info:  &ihfsv1alpha1.FileInfo{Name: "test.txt", Size: 10},
				},
			}

			entries := protofsv1alpha1.FromProtoDirEntries(protos)

			info, err := entries[0].Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Name()).To(Equal("test.txt"))
			Expect(info.Size()).To(Equal(int64(10)))
		})

		It("should return nil Info for proto DirEntry without info", func() {
			protos := []*ihfsv1alpha1.DirEntry{
				{Name: "test.txt", IsDir: false},
			}

			entries := protofsv1alpha1.FromProtoDirEntries(protos)

			info, err := entries[0].Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(info).To(BeNil())
		})

		It("should return empty slice for empty input", func() {
			entries := protofsv1alpha1.FromProtoDirEntries(nil)
			Expect(entries).To(BeEmpty())
		})
	})
})

// errInfoDirEntry is a DirEntry whose Info() always returns an error.
type errInfoDirEntry struct {
	name string
}

func (e *errInfoDirEntry) Name() string               { return e.name }
func (e *errInfoDirEntry) IsDir() bool                { return false }
func (e *errInfoDirEntry) Type() fs.FileMode          { return 0 }
func (e *errInfoDirEntry) Info() (fs.FileInfo, error) { return nil, errors.New("info error") }

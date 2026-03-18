package protofsv1alpha1_test

import (
	"errors"
	"io/fs"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	filev1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/dev/unmango/file/v1alpha1"
	protofsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/grpc/v1alpha1"
	"github.com/unstoppablemango/ihfs/testfs"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ = Describe("DirEntry", func() {
	Describe("DirEntryToFileInfo", func() {
		It("should convert a directory DirEntry to FileInfo", func() {
			entry := testfs.NewDirEntry("subdir", true)

			fi := protofsv1alpha1.DirEntryToFileInfo(entry)

			Expect(fi.Name).To(Equal("subdir"))
			Expect(fi.IsDir).To(BeTrue())
		})

		It("should convert a file DirEntry to FileInfo", func() {
			entry := testfs.NewDirEntry("file.txt", false)

			fi := protofsv1alpha1.DirEntryToFileInfo(entry)

			Expect(fi.Name).To(Equal("file.txt"))
			Expect(fi.IsDir).To(BeFalse())
		})

		It("should handle DirEntry with failing Info()", func() {
			entry := &errInfoDirEntry{name: "broken"}

			fi := protofsv1alpha1.DirEntryToFileInfo(entry)

			Expect(fi.Name).To(Equal("broken"))
			Expect(fi.IsDir).To(BeFalse())
		})
	})

	Describe("FileInfosToDirEntries", func() {
		It("should convert proto FileInfos to DirEntries", func() {
			now := time.Now().UTC().Truncate(time.Second)
			infos := []*filev1alpha1.FileInfo{
				{Name: "dir1", IsDir: true, Mode: filev1alpha1.FileMode_FILE_MODE_DIR},
				{Name: "file.txt", IsDir: false, ModTime: timestamppb.New(now)},
			}

			entries := protofsv1alpha1.FileInfosToDirEntries(infos)

			Expect(entries).To(HaveLen(2))
			Expect(entries[0].Name()).To(Equal("dir1"))
			Expect(entries[0].IsDir()).To(BeTrue())
			Expect(entries[0].Type()).To(Equal(fs.ModeDir))
			Expect(entries[1].Name()).To(Equal("file.txt"))
			Expect(entries[1].IsDir()).To(BeFalse())
		})

		It("should return Info from proto FileInfo", func() {
			infos := []*filev1alpha1.FileInfo{
				{Name: "test.txt", IsDir: false, Size: 10},
			}

			entries := protofsv1alpha1.FileInfosToDirEntries(infos)

			info, err := entries[0].Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Name()).To(Equal("test.txt"))
			Expect(info.Size()).To(Equal(int64(10)))
		})

		It("should return empty slice for empty input", func() {
			entries := protofsv1alpha1.FileInfosToDirEntries(nil)
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

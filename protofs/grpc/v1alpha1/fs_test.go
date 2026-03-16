package protofsv1alpha1_test

import (
	"context"
	"io"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	fsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/ihfs/fs/v1alpha1"
	protofsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/grpc/v1alpha1"
	"github.com/unstoppablemango/ihfs/testfs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// minimalFS only implements ihfs.FS (Open), no extended interfaces.
type minimalFS struct{}

func (minimalFS) Open(string) (ihfs.File, error) {
	return &minimalFile{}, nil
}

// minimalFile implements only fs.File (Read, Close, Stat).
type minimalFile struct{}

func (minimalFile) Read([]byte) (int, error)   { return 0, io.EOF }
func (minimalFile) Close() error               { return nil }
func (minimalFile) Stat() (fs.FileInfo, error) { return testfs.NewFileInfo("test.txt"), nil }

var _ = Describe("FsServer", func() {
	var server *protofsv1alpha1.FsServer

	BeforeEach(func() {
		server = protofsv1alpha1.NewServer(minimalFS{})
	})

	It("should return Unimplemented for Stat", func() {
		_, err := server.Stat(context.Background(), &fsv1alpha1.StatRequest{Name: "f"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for ReadDir", func() {
		_, err := server.ReadDir(context.Background(), &fsv1alpha1.ReadDirRequest{Name: "."})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for ReadFile", func() {
		_, err := server.ReadFile(context.Background(), &fsv1alpha1.ReadFileRequest{Name: "f"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Glob", func() {
		_, err := server.Glob(context.Background(), &fsv1alpha1.GlobRequest{Pattern: "*.txt"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Create", func() {
		_, err := server.Create(context.Background(), &fsv1alpha1.CreateRequest{Name: "f"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for WriteFile", func() {
		_, err := server.WriteFile(context.Background(), &fsv1alpha1.WriteFileRequest{Name: "f"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Mkdir", func() {
		_, err := server.Mkdir(context.Background(), &fsv1alpha1.MkdirRequest{Name: "d"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for MkdirAll", func() {
		_, err := server.MkdirAll(context.Background(), &fsv1alpha1.MkdirAllRequest{Name: "a/b"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Remove", func() {
		_, err := server.Remove(context.Background(), &fsv1alpha1.RemoveRequest{Name: "f"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for RemoveAll", func() {
		_, err := server.RemoveAll(context.Background(), &fsv1alpha1.RemoveAllRequest{Name: "d"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Rename", func() {
		_, err := server.Rename(context.Background(), &fsv1alpha1.RenameRequest{Oldpath: "a", Newpath: "b"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Chmod", func() {
		_, err := server.Chmod(context.Background(), &fsv1alpha1.ChmodRequest{Name: "f"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Chown", func() {
		_, err := server.Chown(context.Background(), &fsv1alpha1.ChownRequest{Name: "f"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Chtimes", func() {
		_, err := server.Chtimes(context.Background(), &fsv1alpha1.ChtimesRequest{Name: "f"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Symlink", func() {
		_, err := server.Symlink(context.Background(), &fsv1alpha1.SymlinkRequest{Oldname: "a", Newname: "b"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for ReadLink", func() {
		_, err := server.ReadLink(context.Background(), &fsv1alpha1.ReadLinkRequest{Name: "l"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Lstat", func() {
		_, err := server.Lstat(context.Background(), &fsv1alpha1.LstatRequest{Name: "l"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})
})

var _ = Describe("FsServer operation errors", func() {
	var server *protofsv1alpha1.FsServer

	BeforeEach(func() {
		server = protofsv1alpha1.NewServer(testfs.New())
	})

	It("should propagate Open errors", func() {
		_, err := server.Open(context.Background(), &fsv1alpha1.OpenRequest{Name: "f"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate ReadDir errors", func() {
		_, err := server.ReadDir(context.Background(), &fsv1alpha1.ReadDirRequest{Name: "."})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate ReadFile errors", func() {
		_, err := server.ReadFile(context.Background(), &fsv1alpha1.ReadFileRequest{Name: "f"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Glob errors", func() {
		_, err := server.Glob(context.Background(), &fsv1alpha1.GlobRequest{Pattern: "*.txt"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Create errors", func() {
		_, err := server.Create(context.Background(), &fsv1alpha1.CreateRequest{Name: "f"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate WriteFile errors", func() {
		_, err := server.WriteFile(context.Background(), &fsv1alpha1.WriteFileRequest{Name: "f"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Mkdir errors", func() {
		_, err := server.Mkdir(context.Background(), &fsv1alpha1.MkdirRequest{Name: "d"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate MkdirAll errors", func() {
		_, err := server.MkdirAll(context.Background(), &fsv1alpha1.MkdirAllRequest{Name: "a/b"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Remove errors", func() {
		_, err := server.Remove(context.Background(), &fsv1alpha1.RemoveRequest{Name: "f"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate RemoveAll errors", func() {
		_, err := server.RemoveAll(context.Background(), &fsv1alpha1.RemoveAllRequest{Name: "d"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Rename errors", func() {
		_, err := server.Rename(context.Background(), &fsv1alpha1.RenameRequest{Oldpath: "a", Newpath: "b"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Chmod errors", func() {
		_, err := server.Chmod(context.Background(), &fsv1alpha1.ChmodRequest{Name: "f"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Chown errors", func() {
		_, err := server.Chown(context.Background(), &fsv1alpha1.ChownRequest{Name: "f"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Chtimes errors", func() {
		_, err := server.Chtimes(context.Background(), &fsv1alpha1.ChtimesRequest{Name: "f"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Symlink errors", func() {
		_, err := server.Symlink(context.Background(), &fsv1alpha1.SymlinkRequest{Oldname: "a", Newname: "b"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate ReadLink errors", func() {
		_, err := server.ReadLink(context.Background(), &fsv1alpha1.ReadLinkRequest{Name: "l"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Lstat errors", func() {
		_, err := server.Lstat(context.Background(), &fsv1alpha1.LstatRequest{Name: "l"})
		Expect(err).To(HaveOccurred())
	})
})

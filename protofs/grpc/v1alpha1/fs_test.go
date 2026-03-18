package protofsv1alpha1_test

import (
	"context"
	"io"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	fsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/dev/unmango/fs/v1alpha1"
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

	It("should return Unimplemented for Create", func() {
		_, err := server.Create(context.Background(), &fsv1alpha1.CreateRequest{Name: "f"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for OpenFile", func() {
		_, err := server.OpenFile(context.Background(), &fsv1alpha1.OpenFileRequest{Name: "f"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Mkdir", func() {
		_, err := server.Mkdir(context.Background(), &fsv1alpha1.MkdirRequest{Name: "d"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for MkdirAll", func() {
		_, err := server.MkdirAll(context.Background(), &fsv1alpha1.MkdirAllRequest{Path: "a/b"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Remove", func() {
		_, err := server.Remove(context.Background(), &fsv1alpha1.RemoveRequest{Name: "f"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for RemoveAll", func() {
		_, err := server.RemoveAll(context.Background(), &fsv1alpha1.RemoveAllRequest{Path: "d"})
		Expect(status.Code(err)).To(Equal(codes.Unimplemented))
	})

	It("should return Unimplemented for Rename", func() {
		_, err := server.Rename(context.Background(), &fsv1alpha1.RenameRequest{Oldname: "a", Newname: "b"})
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

	It("should propagate Create errors", func() {
		_, err := server.Create(context.Background(), &fsv1alpha1.CreateRequest{Name: "f"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Mkdir errors", func() {
		_, err := server.Mkdir(context.Background(), &fsv1alpha1.MkdirRequest{Name: "d"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate MkdirAll errors", func() {
		_, err := server.MkdirAll(context.Background(), &fsv1alpha1.MkdirAllRequest{Path: "a/b"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Remove errors", func() {
		_, err := server.Remove(context.Background(), &fsv1alpha1.RemoveRequest{Name: "f"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate RemoveAll errors", func() {
		_, err := server.RemoveAll(context.Background(), &fsv1alpha1.RemoveAllRequest{Path: "d"})
		Expect(err).To(HaveOccurred())
	})

	It("should propagate Rename errors", func() {
		_, err := server.Rename(context.Background(), &fsv1alpha1.RenameRequest{Oldname: "a", Newname: "b"})
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
})

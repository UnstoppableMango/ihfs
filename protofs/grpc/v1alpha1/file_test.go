package protofsv1alpha1_test

import (
	"context"
	"errors"
	"io"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	filev1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/ihfs/file/v1alpha1"
	ihfsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/ihfs/v1alpha1"
	protofsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/grpc/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fileRef is a convenience helper.
var fileRef = &ihfsv1alpha1.File{Name: "test.txt"}

// createOnlyFS implements CreateFS but Create returns a minimalFile.
type createOnlyFS struct {
	minimalFS
}

func (createOnlyFS) Create(string) (ihfs.File, error) {
	return &minimalFile{}, nil
}

var _ = Describe("FileServer", func() {
	Describe("with minimal FS (no write or dir support)", func() {
		var server *protofsv1alpha1.FileServer

		BeforeEach(func() {
			server = protofsv1alpha1.NewFileServer(minimalFS{})
		})

		It("should return Unimplemented for Write (no CreateFS)", func() {
			_, err := server.Write(context.Background(), &filev1alpha1.WriteRequest{
				File: fileRef,
				Data: []byte("data"),
			})
			Expect(status.Code(err)).To(Equal(codes.Unimplemented))
		})

		It("should return Unimplemented for WriteAt (no CreateFS)", func() {
			_, err := server.WriteAt(context.Background(), &filev1alpha1.WriteAtRequest{
				File:   fileRef,
				Data:   []byte("data"),
				Offset: 0,
			})
			Expect(status.Code(err)).To(Equal(codes.Unimplemented))
		})

		It("should return Unimplemented for ReadDir (file has no ReadDir)", func() {
			_, err := server.ReadDir(context.Background(), &filev1alpha1.ReadDirRequest{
				File: fileRef,
				N:    -1,
			})
			Expect(status.Code(err)).To(Equal(codes.Unimplemented))
		})

		It("should return Unimplemented for Sync (file has no Sync)", func() {
			_, err := server.Sync(context.Background(), &filev1alpha1.SyncRequest{File: fileRef})
			Expect(status.Code(err)).To(Equal(codes.Unimplemented))
		})

		It("should return Unimplemented for Truncate (file has no Truncate)", func() {
			_, err := server.Truncate(context.Background(), &filev1alpha1.TruncateRequest{
				File: fileRef,
				Size: 0,
			})
			Expect(status.Code(err)).To(Equal(codes.Unimplemented))
		})

		It("should succeed for Close", func() {
			_, err := server.Close(context.Background(), &filev1alpha1.CloseRequest{File: fileRef})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("with createOnly FS (CreateFS returns minimalFile without Write/WriteAt)", func() {
		var server *protofsv1alpha1.FileServer

		BeforeEach(func() {
			server = protofsv1alpha1.NewFileServer(createOnlyFS{})
		})

		It("should return Unimplemented for Write (file has no io.Writer)", func() {
			_, err := server.Write(context.Background(), &filev1alpha1.WriteRequest{
				File: fileRef,
				Data: []byte("data"),
			})
			Expect(status.Code(err)).To(Equal(codes.Unimplemented))
		})

		It("should return Unimplemented for WriteAt (file has no io.WriterAt)", func() {
			_, err := server.WriteAt(context.Background(), &filev1alpha1.WriteAtRequest{
				File:   fileRef,
				Data:   []byte("data"),
				Offset: 0,
			})
			Expect(status.Code(err)).To(Equal(codes.Unimplemented))
		})
	})

	Describe("Read", func() {
		It("should read file content", func() {
			content := []byte("test content")
			fsys := &singleFileFS{
				name:    "test.txt",
				content: content,
			}
			server := protofsv1alpha1.NewFileServer(fsys)

			res, err := server.Read(context.Background(), &filev1alpha1.ReadRequest{File: fileRef})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Data).To(Equal(content))
		})
	})

	Describe("Stat", func() {
		It("should return file stat", func() {
			fsys := minimalFS{}
			server := protofsv1alpha1.NewFileServer(fsys)

			res, err := server.Stat(context.Background(), &filev1alpha1.StatRequest{File: fileRef})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.FileInfo).NotTo(BeNil())
		})
	})
})

// singleFileFS returns a file with specific content.
type singleFileFS struct {
	name    string
	content []byte
}

func (f *singleFileFS) Open(string) (ihfs.File, error) {
	return &contentFile{content: f.content}, nil
}

// contentFile is a file with specific readable content.
type contentFile struct {
	content []byte
	pos     int
}

func (f *contentFile) Read(p []byte) (int, error) {
	if f.pos >= len(f.content) {
		return 0, io.EOF
	}

	n := copy(p, f.content[f.pos:])
	f.pos += n

	return n, nil
}

func (f *contentFile) Close() error               { return nil }
func (f *contentFile) Stat() (fs.FileInfo, error) { return nil, nil }

// errorOpenFS always fails to open.
type errorOpenFS struct{}

func (errorOpenFS) Open(string) (ihfs.File, error) { return nil, errors.New("open error") }

// readErrorFile's Read returns a real error (not EOF).
type readErrorFile struct{}

func (readErrorFile) Read([]byte) (int, error)   { return 0, errors.New("read error") }
func (readErrorFile) Close() error               { return nil }
func (readErrorFile) Stat() (fs.FileInfo, error) { return nil, nil }

type readErrorFS struct{}

func (readErrorFS) Open(string) (ihfs.File, error) { return readErrorFile{}, nil }

// statErrorFile's Stat returns an error.
type statErrorFile struct{}

func (statErrorFile) Read([]byte) (int, error)   { return 0, io.EOF }
func (statErrorFile) Close() error               { return nil }
func (statErrorFile) Stat() (fs.FileInfo, error) { return nil, errors.New("stat error") }

type statErrorFS struct{}

func (statErrorFS) Open(string) (ihfs.File, error) { return statErrorFile{}, nil }

// readDirErrorFile implements fs.ReadDirFile but ReadDir returns an error.
type readDirErrorFile struct{}

func (readDirErrorFile) Read([]byte) (int, error)              { return 0, io.EOF }
func (readDirErrorFile) Close() error                          { return nil }
func (readDirErrorFile) Stat() (fs.FileInfo, error)            { return nil, nil }
func (readDirErrorFile) ReadDir(int) ([]fs.DirEntry, error)    { return nil, errors.New("readdir error") }

type readDirErrorFS struct{}

func (readDirErrorFS) Open(string) (ihfs.File, error) { return readDirErrorFile{}, nil }

// errorCreateFS implements CreateFS but Create returns an error.
type errorCreateFS struct{ minimalFS }

func (errorCreateFS) Create(string) (ihfs.File, error) { return nil, errors.New("create error") }

// writeErrorFile implements io.Writer but Write returns an error.
type writeErrorFile struct{}

func (writeErrorFile) Read([]byte) (int, error)   { return 0, io.EOF }
func (writeErrorFile) Close() error               { return nil }
func (writeErrorFile) Stat() (fs.FileInfo, error) { return nil, nil }
func (writeErrorFile) Write([]byte) (int, error)  { return 0, errors.New("write error") }

type writeErrorCreateFS struct{ minimalFS }

func (writeErrorCreateFS) Create(string) (ihfs.File, error) { return writeErrorFile{}, nil }

// writeAtErrorFile implements io.WriterAt but WriteAt returns an error.
type writeAtErrorFile struct{}

func (writeAtErrorFile) Read([]byte) (int, error)           { return 0, io.EOF }
func (writeAtErrorFile) Close() error                       { return nil }
func (writeAtErrorFile) Stat() (fs.FileInfo, error)         { return nil, nil }
func (writeAtErrorFile) WriteAt([]byte, int64) (int, error) { return 0, errors.New("writeat error") }

type writeAtErrorCreateFS struct{ minimalFS }

func (writeAtErrorCreateFS) Create(string) (ihfs.File, error) { return writeAtErrorFile{}, nil }

// syncErrorFile implements ihfs.Syncer but Sync returns an error.
type syncErrorFile struct{}

func (syncErrorFile) Read([]byte) (int, error)   { return 0, io.EOF }
func (syncErrorFile) Close() error               { return nil }
func (syncErrorFile) Stat() (fs.FileInfo, error) { return nil, nil }
func (syncErrorFile) Sync() error                { return errors.New("sync error") }

type syncErrorFS struct{}

func (syncErrorFS) Open(string) (ihfs.File, error) { return syncErrorFile{}, nil }

// truncateErrorFile implements ihfs.Truncater but Truncate returns an error.
type truncateErrorFile struct{}

func (truncateErrorFile) Read([]byte) (int, error)   { return 0, io.EOF }
func (truncateErrorFile) Close() error               { return nil }
func (truncateErrorFile) Stat() (fs.FileInfo, error) { return nil, nil }
func (truncateErrorFile) Truncate(int64) error       { return errors.New("truncate error") }

type truncateErrorFS struct{}

func (truncateErrorFS) Open(string) (ihfs.File, error) { return truncateErrorFile{}, nil }

var _ = Describe("FileServer error paths", func() {
	Describe("Read", func() {
		It("should return error when Open fails", func() {
			server := protofsv1alpha1.NewFileServer(errorOpenFS{})
			_, err := server.Read(context.Background(), &filev1alpha1.ReadRequest{File: fileRef})
			Expect(err).To(HaveOccurred())
		})

		It("should return error when io.ReadAll fails", func() {
			server := protofsv1alpha1.NewFileServer(readErrorFS{})
			_, err := server.Read(context.Background(), &filev1alpha1.ReadRequest{File: fileRef})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Stat", func() {
		It("should return error when Open fails", func() {
			server := protofsv1alpha1.NewFileServer(errorOpenFS{})
			_, err := server.Stat(context.Background(), &filev1alpha1.StatRequest{File: fileRef})
			Expect(err).To(HaveOccurred())
		})

		It("should return error when file.Stat fails", func() {
			server := protofsv1alpha1.NewFileServer(statErrorFS{})
			_, err := server.Stat(context.Background(), &filev1alpha1.StatRequest{File: fileRef})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ReadDir", func() {
		It("should return error when Open fails", func() {
			server := protofsv1alpha1.NewFileServer(errorOpenFS{})
			_, err := server.ReadDir(context.Background(), &filev1alpha1.ReadDirRequest{File: fileRef, N: -1})
			Expect(err).To(HaveOccurred())
		})

		It("should return error when ReadDir fails", func() {
			server := protofsv1alpha1.NewFileServer(readDirErrorFS{})
			_, err := server.ReadDir(context.Background(), &filev1alpha1.ReadDirRequest{File: fileRef, N: -1})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Write", func() {
		It("should return error when Create fails", func() {
			server := protofsv1alpha1.NewFileServer(errorCreateFS{})
			_, err := server.Write(context.Background(), &filev1alpha1.WriteRequest{File: fileRef, Data: []byte("x")})
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Write fails", func() {
			server := protofsv1alpha1.NewFileServer(writeErrorCreateFS{})
			_, err := server.Write(context.Background(), &filev1alpha1.WriteRequest{File: fileRef, Data: []byte("x")})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("WriteAt", func() {
		It("should return error when Create fails", func() {
			server := protofsv1alpha1.NewFileServer(errorCreateFS{})
			_, err := server.WriteAt(context.Background(), &filev1alpha1.WriteAtRequest{File: fileRef, Data: []byte("x")})
			Expect(err).To(HaveOccurred())
		})

		It("should return error when WriteAt fails", func() {
			server := protofsv1alpha1.NewFileServer(writeAtErrorCreateFS{})
			_, err := server.WriteAt(context.Background(), &filev1alpha1.WriteAtRequest{File: fileRef, Data: []byte("x")})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Sync", func() {
		It("should return error when Open fails", func() {
			server := protofsv1alpha1.NewFileServer(errorOpenFS{})
			_, err := server.Sync(context.Background(), &filev1alpha1.SyncRequest{File: fileRef})
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Sync fails", func() {
			server := protofsv1alpha1.NewFileServer(syncErrorFS{})
			_, err := server.Sync(context.Background(), &filev1alpha1.SyncRequest{File: fileRef})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Truncate", func() {
		It("should return error when Open fails", func() {
			server := protofsv1alpha1.NewFileServer(errorOpenFS{})
			_, err := server.Truncate(context.Background(), &filev1alpha1.TruncateRequest{File: fileRef, Size: 0})
			Expect(err).To(HaveOccurred())
		})

		It("should return error when Truncate fails", func() {
			server := protofsv1alpha1.NewFileServer(truncateErrorFS{})
			_, err := server.Truncate(context.Background(), &filev1alpha1.TruncateRequest{File: fileRef, Size: 0})
			Expect(err).To(HaveOccurred())
		})
	})
})

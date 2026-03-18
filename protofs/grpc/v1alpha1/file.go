package protofsv1alpha1

import (
	"context"
	"io"
	"io/fs"

	"github.com/unstoppablemango/ihfs"
	filev1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/dev/unmango/file/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// File is a gRPC client implementing ihfs.File.
// It lazily fetches file content from the server on the first Read call
// and maintains a local cursor for subsequent reads.
type File struct {
	client filev1alpha1.FileServiceClient
	file   *filev1alpha1.File
	ctx    ContextFunc

	// Lazily loaded content and cursor position.
	data   []byte
	pos    int
	loaded bool
}

// Read implements fs.File.
func (f *File) Read(p []byte) (int, error) {
	if err := f.load(); err != nil {
		return 0, err
	}

	if f.pos >= len(f.data) {
		return 0, io.EOF
	}

	n := copy(p, f.data[f.pos:])
	f.pos += n

	return n, nil
}

// Close implements fs.File.
func (f *File) Close() error {
	return nil
}

// Stat implements fs.File.
func (f *File) Stat() (fs.FileInfo, error) {
	res, err := f.client.Stat(f.ctx(), &filev1alpha1.StatRequest{File: f.file})
	if err != nil {
		return nil, err
	}

	return fromProtoFileInfo(res.FileInfo), nil
}

// ReadDir implements ihfs.DirReader (fs.ReadDirFile).
func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {
	res, err := f.client.Readdir(f.ctx(), &filev1alpha1.ReaddirRequest{
		File:  f.file,
		Count: int32(n),
	})
	if err != nil {
		return nil, err
	}

	return fileInfosToDirEntries(res.FileInfos), nil
}

// Write implements ihfs.Writer.
func (f *File) Write(p []byte) (int, error) {
	_, err := f.client.Write(f.ctx(), &filev1alpha1.WriteRequest{
		File: f.file,
		Data: p,
	})
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// WriteAt implements ihfs.WriterAt.
func (f *File) WriteAt(p []byte, off int64) (int, error) {
	_, err := f.client.WriteAt(f.ctx(), &filev1alpha1.WriteAtRequest{
		File:   f.file,
		Data:   p,
		Offset: off,
	})
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// Truncate implements ihfs.Truncater.
func (f *File) Truncate(size int64) error {
	_, err := f.client.Truncate(f.ctx(), &filev1alpha1.TruncateRequest{
		File: f.file,
		Size: size,
	})

	return err
}

// load fetches all file content from the server if not already loaded.
func (f *File) load() error {
	if f.loaded {
		return nil
	}

	res, err := f.client.Read(f.ctx(), &filev1alpha1.ReadRequest{File: f.file})
	if err != nil {
		return err
	}

	f.data = res.Data
	f.loaded = true

	return nil
}

// Verify interface compliance.
var (
	_ ihfs.File      = (*File)(nil)
	_ ihfs.DirReader = (*File)(nil)
	_ ihfs.Writer    = (*File)(nil)
	_ ihfs.WriterAt  = (*File)(nil)
	_ ihfs.Truncater = (*File)(nil)
)

// FileServer wraps an ihfs.FS and implements the gRPC FileService.
type FileServer struct {
	filev1alpha1.UnimplementedFileServiceServer

	fs ihfs.FS
}

// NewFileServer creates a new FileServer wrapping the provided filesystem.
func NewFileServer(fsys ihfs.FS) *FileServer {
	return &FileServer{fs: fsys}
}

// RegisterFileServer registers a FileServer for the provided filesystem with the gRPC server.
func RegisterFileServer(s grpc.ServiceRegistrar, fsys ihfs.FS) {
	filev1alpha1.RegisterFileServiceServer(s, NewFileServer(fsys))
}

func (s *FileServer) Read(_ context.Context, req *filev1alpha1.ReadRequest) (*filev1alpha1.ReadResponse, error) {
	file, err := s.fs.Open(req.File.Name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return &filev1alpha1.ReadResponse{Data: data}, nil
}

func (s *FileServer) ReadAt(_ context.Context, req *filev1alpha1.ReadAtRequest) (*filev1alpha1.ReadAtResponse, error) {
	file, err := s.fs.Open(req.File.Name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	offset := req.Offset
	if offset < 0 || offset > int64(len(data)) {
		return nil, status.Error(codes.OutOfRange, "offset out of range")
	}

	return &filev1alpha1.ReadAtResponse{Data: data[offset:]}, nil
}

func (s *FileServer) Readdir(_ context.Context, req *filev1alpha1.ReaddirRequest) (*filev1alpha1.ReaddirResponse, error) {
	if file, err := s.fs.Open(req.File.Name); err == nil {
		defer func() { _ = file.Close() }()

		if rdf, ok := file.(fs.ReadDirFile); ok {
			entries, err := rdf.ReadDir(int(req.Count))
			if err != nil {
				return nil, err
			}

			infos := make([]*filev1alpha1.FileInfo, len(entries))
			for i, e := range entries {
				infos[i] = dirEntryToFileInfo(e)
			}

			return &filev1alpha1.ReaddirResponse{FileInfos: infos}, nil
		}
	}

	rdfs, ok := s.fs.(ihfs.ReadDirFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "Readdir not supported")
	}

	entries, err := rdfs.ReadDir(req.File.Name)
	if err != nil {
		return nil, err
	}

	infos := make([]*filev1alpha1.FileInfo, len(entries))
	for i, e := range entries {
		infos[i] = dirEntryToFileInfo(e)
	}

	return &filev1alpha1.ReaddirResponse{FileInfos: infos}, nil
}

func (s *FileServer) ReaddirNames(_ context.Context, req *filev1alpha1.ReaddirNamesRequest) (*filev1alpha1.ReaddirNamesResponse, error) {
	file, err := s.fs.Open(req.File.Name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	rdf, ok := file.(fs.ReadDirFile)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "ReaddirNames not supported by this file")
	}

	entries, err := rdf.ReadDir(int(req.Count))
	if err != nil {
		return nil, err
	}

	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}

	return &filev1alpha1.ReaddirNamesResponse{Names: names}, nil
}

func (s *FileServer) Stat(_ context.Context, req *filev1alpha1.StatRequest) (*filev1alpha1.StatResponse, error) {
	file, err := s.fs.Open(req.File.Name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return &filev1alpha1.StatResponse{FileInfo: toProtoFileInfo(info)}, nil
}

func (s *FileServer) Truncate(_ context.Context, req *filev1alpha1.TruncateRequest) (*filev1alpha1.TruncateResponse, error) {
	file, err := s.fs.Open(req.File.Name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	t, ok := file.(ihfs.Truncater)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "Truncate not supported by this file")
	}

	if err := t.Truncate(req.Size); err != nil {
		return nil, err
	}

	return &filev1alpha1.TruncateResponse{}, nil
}

func (s *FileServer) Write(_ context.Context, req *filev1alpha1.WriteRequest) (*filev1alpha1.WriteResponse, error) {
	fsys, ok := s.fs.(ihfs.CreateFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "Write not supported")
	}

	file, err := fsys.Create(req.File.Name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	w, ok := file.(io.Writer)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "Write not supported by this file")
	}

	if _, err := w.Write(req.Data); err != nil {
		return nil, err
	}

	return &filev1alpha1.WriteResponse{}, nil
}

func (s *FileServer) WriteAt(_ context.Context, req *filev1alpha1.WriteAtRequest) (*filev1alpha1.WriteAtResponse, error) {
	fsys, ok := s.fs.(ihfs.CreateFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "WriteAt not supported")
	}

	file, err := fsys.Create(req.File.Name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	wa, ok := file.(io.WriterAt)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "WriteAt not supported by this file")
	}

	if _, err := wa.WriteAt(req.Data, req.Offset); err != nil {
		return nil, err
	}

	return &filev1alpha1.WriteAtResponse{}, nil
}

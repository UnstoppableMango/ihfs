// Package protofsv1alpha1 provides gRPC client and server implementations of ihfs.FS.
package protofsv1alpha1

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"time"

	"github.com/unstoppablemango/ihfs"
	filev1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/dev/unmango/file/v1alpha1"
	fsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/gen/dev/unmango/fs/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ContextFunc returns a context for use in a single RPC call.
type ContextFunc func() context.Context

// Fs is a gRPC client implementing ihfs.FS and additional write interfaces.
type Fs struct {
	fs   fsv1alpha1.FsServiceClient
	file filev1alpha1.FileServiceClient
	ctx  ContextFunc
}

// New creates an Fs client using the provided gRPC connection.
func New(conn grpc.ClientConnInterface) *Fs {
	return NewWithContext(conn, func() context.Context { return context.Background() })
}

// NewWithContext creates an Fs client with a custom context factory.
func NewWithContext(conn grpc.ClientConnInterface, ctx ContextFunc) *Fs {
	return &Fs{
		fs:   fsv1alpha1.NewFsServiceClient(conn),
		file: filev1alpha1.NewFileServiceClient(conn),
		ctx:  ctx,
	}
}

// Open implements ihfs.FS.
func (f *Fs) Open(name string) (ihfs.File, error) {
	res, err := f.fs.Open(f.ctx(), &fsv1alpha1.OpenRequest{Name: name})
	if err != nil {
		return nil, err
	}

	return &File{client: f.file, file: res.File, ctx: f.ctx}, nil
}

// Stat implements ihfs.StatFS.
func (f *Fs) Stat(name string) (ihfs.FileInfo, error) {
	res, err := f.fs.Stat(f.ctx(), &fsv1alpha1.StatRequest{Name: name})
	if err != nil {
		return nil, err
	}

	return fromProtoFileInfo(res.FileInfo), nil
}

// ReadDir implements ihfs.ReadDirFS.
func (f *Fs) ReadDir(name string) ([]ihfs.DirEntry, error) {
	file := &filev1alpha1.File{Name: name}
	res, err := f.file.Readdir(f.ctx(), &filev1alpha1.ReaddirRequest{
		File:  file,
		Count: -1,
	})
	if err != nil {
		return nil, err
	}

	return fileInfosToDirEntries(res.FileInfos), nil
}

// ReadFile implements ihfs.ReadFileFS.
func (f *Fs) ReadFile(name string) ([]byte, error) {
	stream, err := f.fs.Read(f.ctx(), &fsv1alpha1.ReadRequest{Name: name})
	if err != nil {
		return nil, err
	}

	var data []byte
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		data = append(data, resp.Data...)
	}

	return data, nil
}

// Create implements ihfs.CreateFS.
func (f *Fs) Create(name string) (ihfs.File, error) {
	res, err := f.fs.Create(f.ctx(), &fsv1alpha1.CreateRequest{Name: name})
	if err != nil {
		return nil, err
	}

	return &File{client: f.file, file: res.File, ctx: f.ctx}, nil
}

// WriteFile implements ihfs.WriteFileFS.
func (f *Fs) WriteFile(name string, data []byte, _ ihfs.FileMode) error {
	stream, err := f.fs.Write(f.ctx())
	if err != nil {
		return err
	}

	if err := stream.Send(&fsv1alpha1.WriteRequest{Name: name, Data: data}); err != nil {
		return err
	}

	_, err = stream.CloseAndRecv()
	return err
}

// Mkdir implements ihfs.MkdirFS.
func (f *Fs) Mkdir(name string, mode ihfs.FileMode) error {
	_, err := f.fs.Mkdir(f.ctx(), &fsv1alpha1.MkdirRequest{
		Name: name,
		Perm: filev1alpha1.FileMode(mode),
	})

	return err
}

// MkdirAll implements ihfs.MkdirAllFS.
func (f *Fs) MkdirAll(name string, mode ihfs.FileMode) error {
	_, err := f.fs.MkdirAll(f.ctx(), &fsv1alpha1.MkdirAllRequest{
		Path: name,
		Perm: filev1alpha1.FileMode(mode),
	})

	return err
}

// Remove implements ihfs.RemoveFS.
func (f *Fs) Remove(name string) error {
	_, err := f.fs.Remove(f.ctx(), &fsv1alpha1.RemoveRequest{Name: name})

	return err
}

// RemoveAll implements ihfs.RemoveAllFS.
func (f *Fs) RemoveAll(name string) error {
	_, err := f.fs.RemoveAll(f.ctx(), &fsv1alpha1.RemoveAllRequest{Path: name})

	return err
}

// Rename implements ihfs.RenameFS.
func (f *Fs) Rename(oldpath, newpath string) error {
	_, err := f.fs.Rename(f.ctx(), &fsv1alpha1.RenameRequest{
		Oldname: oldpath,
		Newname: newpath,
	})

	return err
}

// Chmod implements ihfs.ChmodFS.
func (f *Fs) Chmod(name string, mode ihfs.FileMode) error {
	_, err := f.fs.Chmod(f.ctx(), &fsv1alpha1.ChmodRequest{
		Name: name,
		Mode: filev1alpha1.FileMode(mode),
	})

	return err
}

// Chown implements ihfs.ChownFS.
func (f *Fs) Chown(name string, uid, gid int) error {
	_, err := f.fs.Chown(f.ctx(), &fsv1alpha1.ChownRequest{
		Name: name,
		Uid:  int32(uid),
		Gid:  int32(gid),
	})

	return err
}

// Chtimes implements ihfs.ChtimesFS.
func (f *Fs) Chtimes(name string, atime, mtime time.Time) error {
	_, err := f.fs.Chtimes(f.ctx(), &fsv1alpha1.ChtimesRequest{
		Name:  name,
		Atime: timestamppb.New(atime),
		Mtime: timestamppb.New(mtime),
	})

	return err
}

// OpenFile implements ihfs.OpenFileFS.
func (f *Fs) OpenFile(name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
	res, err := f.fs.OpenFile(f.ctx(), &fsv1alpha1.OpenFileRequest{
		Name: name,
		Flag: int64(flag),
		Perm: filev1alpha1.FileMode(perm),
	})
	if err != nil {
		return nil, err
	}

	return &File{client: f.file, file: res.File, ctx: f.ctx}, nil
}

// Verify interface compliance.
var (
	_ ihfs.FS          = (*Fs)(nil)
	_ ihfs.StatFS      = (*Fs)(nil)
	_ ihfs.ReadDirFS   = (*Fs)(nil)
	_ ihfs.ReadFileFS  = (*Fs)(nil)
	_ ihfs.CreateFS    = (*Fs)(nil)
	_ ihfs.WriteFileFS = (*Fs)(nil)
	_ ihfs.MkdirFS     = (*Fs)(nil)
	_ ihfs.MkdirAllFS  = (*Fs)(nil)
	_ ihfs.RemoveFS    = (*Fs)(nil)
	_ ihfs.RemoveAllFS = (*Fs)(nil)
	_ ihfs.RenameFS    = (*Fs)(nil)
	_ ihfs.ChmodFS     = (*Fs)(nil)
	_ ihfs.ChownFS     = (*Fs)(nil)
	_ ihfs.ChtimesFS   = (*Fs)(nil)
	_ ihfs.OpenFileFS  = (*Fs)(nil)
)

// FsServer wraps an ihfs.FS and implements the gRPC FsService.
type FsServer struct {
	fsv1alpha1.UnimplementedFsServiceServer

	fs ihfs.FS
}

// NewServer creates a new FsServer wrapping the provided filesystem.
func NewServer(fsys ihfs.FS) *FsServer {
	return &FsServer{fs: fsys}
}

// RegisterFsServer registers a FsServer for the provided filesystem with the gRPC server.
func RegisterFsServer(s grpc.ServiceRegistrar, fsys ihfs.FS) {
	fsv1alpha1.RegisterFsServiceServer(s, NewServer(fsys))
}

func (s *FsServer) Open(_ context.Context, req *fsv1alpha1.OpenRequest) (*fsv1alpha1.OpenResponse, error) {
	file, err := s.fs.Open(req.Name)
	if err != nil {
		return nil, err
	}
	_ = file.Close()

	return &fsv1alpha1.OpenResponse{
		File: &filev1alpha1.File{Name: req.Name},
	}, nil
}

func (s *FsServer) OpenFile(_ context.Context, req *fsv1alpha1.OpenFileRequest) (*fsv1alpha1.OpenFileResponse, error) {
	fsys, ok := s.fs.(ihfs.OpenFileFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "OpenFile not supported")
	}

	file, err := fsys.OpenFile(req.Name, int(req.Flag), fs.FileMode(req.Perm))
	if err != nil {
		return nil, err
	}
	_ = file.Close()

	return &fsv1alpha1.OpenFileResponse{
		File: &filev1alpha1.File{Name: req.Name},
	}, nil
}

func (s *FsServer) Stat(_ context.Context, req *fsv1alpha1.StatRequest) (*fsv1alpha1.StatResponse, error) {
	fsys, ok := s.fs.(ihfs.StatFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "Stat not supported")
	}

	info, err := fsys.Stat(req.Name)
	if err != nil {
		return nil, err
	}

	return &fsv1alpha1.StatResponse{FileInfo: toProtoFileInfo(info)}, nil
}

func (s *FsServer) Read(req *fsv1alpha1.ReadRequest, stream grpc.ServerStreamingServer[fsv1alpha1.ReadResponse]) error {
	var data []byte

	if rffs, ok := s.fs.(ihfs.ReadFileFS); ok {
		d, err := rffs.ReadFile(req.Name)
		if err != nil {
			return err
		}
		data = d
	} else {
		f, err := s.fs.Open(req.Name)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		d, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		data = d
	}

	return stream.Send(&fsv1alpha1.ReadResponse{Data: data})
}

func (s *FsServer) Write(stream grpc.ClientStreamingServer[fsv1alpha1.WriteRequest, fsv1alpha1.WriteResponse]) error {
	fsys, ok := s.fs.(ihfs.WriteFileFS)
	if !ok {
		return status.Error(codes.Unimplemented, "Write not supported")
	}

	var name string
	var data []byte
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if name == "" {
			name = req.Name
		}
		data = append(data, req.Data...)
	}

	if err := fsys.WriteFile(name, data, 0); err != nil {
		return err
	}

	return stream.SendAndClose(&fsv1alpha1.WriteResponse{})
}

func (s *FsServer) Create(_ context.Context, req *fsv1alpha1.CreateRequest) (*fsv1alpha1.CreateResponse, error) {
	fsys, ok := s.fs.(ihfs.CreateFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "Create not supported")
	}

	file, err := fsys.Create(req.Name)
	if err != nil {
		return nil, err
	}
	_ = file.Close()

	return &fsv1alpha1.CreateResponse{
		File: &filev1alpha1.File{Name: req.Name},
	}, nil
}

func (s *FsServer) Mkdir(_ context.Context, req *fsv1alpha1.MkdirRequest) (*fsv1alpha1.MkdirResponse, error) {
	fsys, ok := s.fs.(ihfs.MkdirFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "Mkdir not supported")
	}

	if err := fsys.Mkdir(req.Name, fs.FileMode(req.Perm)); err != nil {
		return nil, err
	}

	return &fsv1alpha1.MkdirResponse{}, nil
}

func (s *FsServer) MkdirAll(_ context.Context, req *fsv1alpha1.MkdirAllRequest) (*fsv1alpha1.MkdirAllResponse, error) {
	fsys, ok := s.fs.(ihfs.MkdirAllFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "MkdirAll not supported")
	}

	if err := fsys.MkdirAll(req.Path, fs.FileMode(req.Perm)); err != nil {
		return nil, err
	}

	return &fsv1alpha1.MkdirAllResponse{}, nil
}

func (s *FsServer) Remove(_ context.Context, req *fsv1alpha1.RemoveRequest) (*fsv1alpha1.RemoveResponse, error) {
	fsys, ok := s.fs.(ihfs.RemoveFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "Remove not supported")
	}

	if err := fsys.Remove(req.Name); err != nil {
		return nil, err
	}

	return &fsv1alpha1.RemoveResponse{}, nil
}

func (s *FsServer) RemoveAll(_ context.Context, req *fsv1alpha1.RemoveAllRequest) (*fsv1alpha1.RemoveAllResponse, error) {
	fsys, ok := s.fs.(ihfs.RemoveAllFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "RemoveAll not supported")
	}

	if err := fsys.RemoveAll(req.Path); err != nil {
		return nil, err
	}

	return &fsv1alpha1.RemoveAllResponse{}, nil
}

func (s *FsServer) Rename(_ context.Context, req *fsv1alpha1.RenameRequest) (*fsv1alpha1.RenameResponse, error) {
	fsys, ok := s.fs.(ihfs.RenameFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "Rename not supported")
	}

	if err := fsys.Rename(req.Oldname, req.Newname); err != nil {
		return nil, err
	}

	return &fsv1alpha1.RenameResponse{}, nil
}

func (s *FsServer) Chmod(_ context.Context, req *fsv1alpha1.ChmodRequest) (*fsv1alpha1.ChmodResponse, error) {
	fsys, ok := s.fs.(ihfs.ChmodFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "Chmod not supported")
	}

	if err := fsys.Chmod(req.Name, fs.FileMode(req.Mode)); err != nil {
		return nil, err
	}

	return &fsv1alpha1.ChmodResponse{}, nil
}

func (s *FsServer) Chown(_ context.Context, req *fsv1alpha1.ChownRequest) (*fsv1alpha1.ChownResponse, error) {
	fsys, ok := s.fs.(ihfs.ChownFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "Chown not supported")
	}

	if err := fsys.Chown(req.Name, int(req.Uid), int(req.Gid)); err != nil {
		return nil, err
	}

	return &fsv1alpha1.ChownResponse{}, nil
}

func (s *FsServer) Chtimes(_ context.Context, req *fsv1alpha1.ChtimesRequest) (*fsv1alpha1.ChtimesResponse, error) {
	fsys, ok := s.fs.(ihfs.ChtimesFS)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "Chtimes not supported")
	}

	if err := fsys.Chtimes(req.Name, req.Atime.AsTime(), req.Mtime.AsTime()); err != nil {
		return nil, err
	}

	return &fsv1alpha1.ChtimesResponse{}, nil
}

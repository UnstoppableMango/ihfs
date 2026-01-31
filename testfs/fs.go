package testfs

import (
	"io/fs"
	"testing/fstest"
	"time"

	"github.com/unmango/go/fopt"
	"github.com/unstoppablemango/ihfs"
)

// MapFS is an alias for fstest.MapFS.
type MapFS = fstest.MapFS

// TODO: make this API less awkward to use in tests

type Fs struct {
	OpenFunc         func(string) (ihfs.File, error)
	StatFunc         func(string) (ihfs.FileInfo, error)
	CreateFunc       func(string) (ihfs.File, error)
	CreateTempFunc   func(string, string) (ihfs.File, error)
	WriteFileFunc    func(string, []byte, ihfs.FileMode) error
	ReadFileFunc     func(string) ([]byte, error)
	ChmodFunc        func(string, ihfs.FileMode) error
	ChownFunc        func(string, int, int) error
	ChtimesFunc      func(string, time.Time, time.Time) error
	CopyFunc         func(string, ihfs.FS) error
	GlobFunc         func(string) ([]string, error)
	LstatFunc        func(string) (ihfs.FileInfo, error)
	MkdirFunc        func(string, ihfs.FileMode) error
	MkdirAllFunc     func(string, ihfs.FileMode) error
	MkdirTempFunc    func(string, string) (string, error)
	OpenFileFunc     func(string, int, ihfs.FileMode) (ihfs.File, error)
	ReadDirFunc      func(string) ([]ihfs.DirEntry, error)
	ReadDirNamesFunc func(string) ([]string, error)
	ReadLinkFunc     func(string) (string, error)
	RemoveFunc       func(string) error
	RemoveAllFunc    func(string) error
	RenameFunc       func(string, string) error
	SubFunc          func(string) (ihfs.FS, error)
	SymlinkFunc      func(string, string) error
	TempFileFunc     func(string, string) (string, error)
}

func New(opts ...Option) Fs {
	fs := Fs{
		OpenFunc:         defaultOpenFunc,
		StatFunc:         defaultStatFunc,
		CreateFunc:       defaultCreateFunc,
		CreateTempFunc:   defaultCreateTempFunc,
		WriteFileFunc:    defaultWriteFileFunc,
		ReadFileFunc:     defaultReadFileFunc,
		ChmodFunc:        defaultChmodFunc,
		ChownFunc:        defaultChownFunc,
		ChtimesFunc:      defaultChtimesFunc,
		CopyFunc:         defaultCopyFunc,
		GlobFunc:         defaultGlobFunc,
		LstatFunc:        defaultLstatFunc,
		MkdirFunc:        defaultMkdirFunc,
		MkdirAllFunc:     defaultMkdirAllFunc,
		MkdirTempFunc:    defaultMkdirTempFunc,
		OpenFileFunc:     defaultOpenFileFunc,
		ReadDirFunc:      defaultReadDirFunc,
		ReadDirNamesFunc: defaultReadDirNamesFunc,
		ReadLinkFunc:     defaultReadLinkFunc,
		RemoveFunc:       defaultRemoveFunc,
		RemoveAllFunc:    defaultRemoveAllFunc,
		RenameFunc:       defaultRenameFunc,
		SubFunc:          defaultSubFunc,
		SymlinkFunc:      defaultSymlinkFunc,
		TempFileFunc:     defaultTempFileFunc,
	}

	fopt.ApplyAll(&fs, opts)
	return fs
}

func (fs Fs) Open(name string) (ihfs.File, error) {
	return fs.OpenFunc(name)
}

func defaultOpenFunc(name string) (ihfs.File, error) {
	return nil, fs.ErrNotExist
}

func (fs Fs) Stat(name string) (ihfs.FileInfo, error) {
	return fs.StatFunc(name)
}

func defaultStatFunc(name string) (ihfs.FileInfo, error) {
	return nil, fs.ErrNotExist
}

func (fs Fs) Create(name string) (ihfs.File, error) {
	return fs.CreateFunc(name)
}

func defaultCreateFunc(name string) (ihfs.File, error) {
	return nil, fs.ErrPermission
}

func (fs Fs) WriteFile(name string, data []byte, perm ihfs.FileMode) error {
	return fs.WriteFileFunc(name, data, perm)
}

func defaultWriteFileFunc(name string, data []byte, perm ihfs.FileMode) error {
	return fs.ErrPermission
}

func (fs Fs) Chmod(name string, mode ihfs.FileMode) error {
	return fs.ChmodFunc(name, mode)
}

func defaultChmodFunc(name string, mode ihfs.FileMode) error {
	return fs.ErrPermission
}

func (fs Fs) Chown(name string, uid, gid int) error {
	return fs.ChownFunc(name, uid, gid)
}

func defaultChownFunc(name string, uid, gid int) error {
	return fs.ErrPermission
}

func (fs Fs) Chtimes(name string, atime, mtime time.Time) error {
	return fs.ChtimesFunc(name, atime, mtime)
}

func defaultChtimesFunc(name string, atime, mtime time.Time) error {
	return fs.ErrPermission
}

func (fs Fs) Copy(dir string, src ihfs.FS) error {
	return fs.CopyFunc(dir, src)
}

func defaultCopyFunc(dir string, src ihfs.FS) error {
	return fs.ErrPermission
}

func (fs Fs) Mkdir(name string, mode ihfs.FileMode) error {
	return fs.MkdirFunc(name, mode)
}

func defaultMkdirFunc(name string, mode ihfs.FileMode) error {
	return fs.ErrPermission
}

func (fs Fs) MkdirAll(name string, mode ihfs.FileMode) error {
	return fs.MkdirAllFunc(name, mode)
}

func defaultMkdirAllFunc(name string, mode ihfs.FileMode) error {
	return fs.ErrPermission
}

func (fs Fs) MkdirTemp(dir, pattern string) (string, error) {
	return fs.MkdirTempFunc(dir, pattern)
}

func defaultMkdirTempFunc(dir, pattern string) (string, error) {
	return "", fs.ErrPermission
}

func (fs Fs) Remove(name string) error {
	return fs.RemoveFunc(name)
}

func defaultRemoveFunc(name string) error {
	return fs.ErrPermission
}

func (fs Fs) RemoveAll(name string) error {
	return fs.RemoveAllFunc(name)
}

func defaultRemoveAllFunc(name string) error {
	return fs.ErrPermission
}

func (fs Fs) ReadDir(name string) ([]ihfs.DirEntry, error) {
	return fs.ReadDirFunc(name)
}

func defaultReadDirFunc(name string) ([]ihfs.DirEntry, error) {
	return nil, fs.ErrNotExist
}

func (fs Fs) CreateTemp(dir, pattern string) (ihfs.File, error) {
	return fs.CreateTempFunc(dir, pattern)
}

func defaultCreateTempFunc(dir, pattern string) (ihfs.File, error) {
	return nil, fs.ErrPermission
}

func (fs Fs) Glob(pattern string) ([]string, error) {
	return fs.GlobFunc(pattern)
}

func defaultGlobFunc(pattern string) ([]string, error) {
	return nil, fs.ErrPermission
}

func (fs Fs) Lstat(name string) (ihfs.FileInfo, error) {
	return fs.LstatFunc(name)
}

func defaultLstatFunc(name string) (ihfs.FileInfo, error) {
	return nil, fs.ErrNotExist
}

func (fs Fs) OpenFile(name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
	return fs.OpenFileFunc(name, flag, perm)
}

func defaultOpenFileFunc(name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
	return nil, fs.ErrPermission
}

func (fs Fs) ReadDirNames(name string) ([]string, error) {
	return fs.ReadDirNamesFunc(name)
}

func defaultReadDirNamesFunc(name string) ([]string, error) {
	return nil, fs.ErrNotExist
}

func (fs Fs) ReadFile(name string) ([]byte, error) {
	return fs.ReadFileFunc(name)
}

func defaultReadFileFunc(name string) ([]byte, error) {
	return nil, fs.ErrPermission
}

func (fs Fs) ReadLink(name string) (string, error) {
	return fs.ReadLinkFunc(name)
}

func defaultReadLinkFunc(name string) (string, error) {
	return "", fs.ErrInvalid
}

func (fs Fs) Rename(oldpath, newpath string) error {
	return fs.RenameFunc(oldpath, newpath)
}

func defaultRenameFunc(oldpath, newpath string) error {
	return fs.ErrPermission
}

func (fs Fs) Sub(dir string) (ihfs.FS, error) {
	return fs.SubFunc(dir)
}

func defaultSubFunc(dir string) (ihfs.FS, error) {
	return nil, fs.ErrNotExist
}

func (fs Fs) Symlink(oldname, newname string) error {
	return fs.SymlinkFunc(oldname, newname)
}

func defaultSymlinkFunc(oldname, newname string) error {
	return fs.ErrPermission
}

func (fs Fs) TempFile(dir, pattern string) (string, error) {
	return fs.TempFileFunc(dir, pattern)
}

func defaultTempFileFunc(dir, pattern string) (string, error) {
	return "", fs.ErrPermission
}

package testfs

import (
	"io/fs"
	"testing/fstest"
	"time"

	"github.com/unmango/go/fopt"
	"github.com/unstoppablemango/ihfs"
)

type (
	MapFS   = fstest.MapFS
	MapFile = fstest.MapFile
)

// TODO: make this API less awkward to use in tests

// Fs is a configurable test filesystem implementation.
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

// New creates a new test [Fs] with the given options.
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

// Open implements [ihfs.FS].
func (fs Fs) Open(name string) (ihfs.File, error) {
	return fs.OpenFunc(name)
}

func defaultOpenFunc(_ string) (ihfs.File, error) {
	return nil, fs.ErrNotExist
}

// Stat implements [ihfs.StatFS].
func (fs Fs) Stat(name string) (ihfs.FileInfo, error) {
	return fs.StatFunc(name)
}

func defaultStatFunc(_ string) (ihfs.FileInfo, error) {
	return nil, fs.ErrNotExist
}

// Create implements [ihfs.CreateFS].
func (fs Fs) Create(name string) (ihfs.File, error) {
	return fs.CreateFunc(name)
}

func defaultCreateFunc(_ string) (ihfs.File, error) {
	return nil, fs.ErrPermission
}

// WriteFile implements [ihfs.WriteFileFS].
func (fs Fs) WriteFile(name string, data []byte, perm ihfs.FileMode) error {
	return fs.WriteFileFunc(name, data, perm)
}

func defaultWriteFileFunc(_ string, _ []byte, _ ihfs.FileMode) error {
	return fs.ErrPermission
}

// Chmod implements [ihfs.ChmodFS].
func (fs Fs) Chmod(name string, mode ihfs.FileMode) error {
	return fs.ChmodFunc(name, mode)
}

func defaultChmodFunc(_ string, _ ihfs.FileMode) error {
	return fs.ErrPermission
}

// Chown implements [ihfs.ChownFS].
func (fs Fs) Chown(name string, uid, gid int) error {
	return fs.ChownFunc(name, uid, gid)
}

func defaultChownFunc(_ string, _, _ int) error {
	return fs.ErrPermission
}

// Chtimes implements [ihfs.ChtimesFS].
func (fs Fs) Chtimes(name string, atime, mtime time.Time) error {
	return fs.ChtimesFunc(name, atime, mtime)
}

func defaultChtimesFunc(_ string, _, _ time.Time) error {
	return fs.ErrPermission
}

// Copy implements [ihfs.CopyFS].
func (fs Fs) Copy(dir string, src ihfs.FS) error {
	return fs.CopyFunc(dir, src)
}

func defaultCopyFunc(_ string, _ ihfs.FS) error {
	return fs.ErrPermission
}

// Mkdir implements [ihfs.MkdirFS].
func (fs Fs) Mkdir(name string, mode ihfs.FileMode) error {
	return fs.MkdirFunc(name, mode)
}

func defaultMkdirFunc(_ string, _ ihfs.FileMode) error {
	return fs.ErrPermission
}

// MkdirAll implements [ihfs.MkdirAllFS].
func (fs Fs) MkdirAll(name string, mode ihfs.FileMode) error {
	return fs.MkdirAllFunc(name, mode)
}

func defaultMkdirAllFunc(_ string, _ ihfs.FileMode) error {
	return fs.ErrPermission
}

// MkdirTemp implements [ihfs.MkdirTempFS].
func (fs Fs) MkdirTemp(dir, pattern string) (string, error) {
	return fs.MkdirTempFunc(dir, pattern)
}

func defaultMkdirTempFunc(_, _ string) (string, error) {
	return "", fs.ErrPermission
}

// Remove implements [ihfs.RemoveFS].
func (fs Fs) Remove(name string) error {
	return fs.RemoveFunc(name)
}

func defaultRemoveFunc(_ string) error {
	return fs.ErrPermission
}

// RemoveAll implements [ihfs.RemoveAllFS].
func (fs Fs) RemoveAll(name string) error {
	return fs.RemoveAllFunc(name)
}

func defaultRemoveAllFunc(_ string) error {
	return fs.ErrPermission
}

// ReadDir implements [ihfs.ReadDirFS].
func (fs Fs) ReadDir(name string) ([]ihfs.DirEntry, error) {
	return fs.ReadDirFunc(name)
}

func defaultReadDirFunc(_ string) ([]ihfs.DirEntry, error) {
	return nil, fs.ErrNotExist
}

// CreateTemp implements [ihfs.CreateTempFS].
func (fs Fs) CreateTemp(dir, pattern string) (ihfs.File, error) {
	return fs.CreateTempFunc(dir, pattern)
}

func defaultCreateTempFunc(_, _ string) (ihfs.File, error) {
	return nil, fs.ErrPermission
}

// Glob implements [ihfs.GlobFS].
func (fs Fs) Glob(pattern string) ([]string, error) {
	return fs.GlobFunc(pattern)
}

func defaultGlobFunc(_ string) ([]string, error) {
	return nil, fs.ErrPermission
}

// Lstat implements [ihfs.StatFS] variant for symlinks.
func (fs Fs) Lstat(name string) (ihfs.FileInfo, error) {
	return fs.LstatFunc(name)
}

func defaultLstatFunc(_ string) (ihfs.FileInfo, error) {
	return nil, fs.ErrNotExist
}

// OpenFile implements [ihfs.OpenFileFS].
func (fs Fs) OpenFile(name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
	return fs.OpenFileFunc(name, flag, perm)
}

func defaultOpenFileFunc(_ string, _ int, _ ihfs.FileMode) (ihfs.File, error) {
	return nil, fs.ErrPermission
}

// ReadDirNames implements [ihfs.ReadDirNamesFS].
func (fs Fs) ReadDirNames(name string) ([]string, error) {
	return fs.ReadDirNamesFunc(name)
}

func defaultReadDirNamesFunc(_ string) ([]string, error) {
	return nil, fs.ErrNotExist
}

// ReadFile implements [ihfs.ReadFileFS].
func (fs Fs) ReadFile(name string) ([]byte, error) {
	return fs.ReadFileFunc(name)
}

func defaultReadFileFunc(_ string) ([]byte, error) {
	return nil, fs.ErrPermission
}

// ReadLink implements [ihfs.ReadLinkFS].
func (fs Fs) ReadLink(name string) (string, error) {
	return fs.ReadLinkFunc(name)
}

func defaultReadLinkFunc(_ string) (string, error) {
	return "", fs.ErrInvalid
}

// Rename implements [ihfs.RenameFS].
func (fs Fs) Rename(oldpath, newpath string) error {
	return fs.RenameFunc(oldpath, newpath)
}

func defaultRenameFunc(_, _ string) error {
	return fs.ErrPermission
}

// Sub implements [ihfs.SubFS].
func (fs Fs) Sub(dir string) (ihfs.FS, error) {
	return fs.SubFunc(dir)
}

func defaultSubFunc(_ string) (ihfs.FS, error) {
	return nil, fs.ErrNotExist
}

// Symlink implements [ihfs.SymlinkFS].
func (fs Fs) Symlink(oldname, newname string) error {
	return fs.SymlinkFunc(oldname, newname)
}

func defaultSymlinkFunc(_, _ string) error {
	return fs.ErrPermission
}

// TempFile implements [ihfs.TempFileFS].
func (fs Fs) TempFile(dir, pattern string) (string, error) {
	return fs.TempFileFunc(dir, pattern)
}

func defaultTempFileFunc(_, _ string) (string, error) {
	return "", fs.ErrPermission
}

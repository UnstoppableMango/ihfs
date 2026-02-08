// Package factory provides a factory-pattern test filesystem for use in unit tests.
package factory

import (
	"errors"
	"time"

	"github.com/unstoppablemango/ihfs"
)

type (
	// OpenFunc is a function that opens a file by name.
	OpenFunc func(string) (ihfs.File, error)
	// StatFunc is a function that returns file info for the named file.
	StatFunc func(string) (ihfs.FileInfo, error)
)

// ErrNotMocked is returned when an operation has no mock configured.
var ErrNotMocked = errors.New("operation has no mock")

// Fs is a test filesystem that returns pre-configured responses in sequence.
type Fs struct {
	name         string
	open         []OpenFunc
	stat         []StatFunc
	chmod        []ChmodFunc
	chown        []ChownFunc
	chtimes      []ChtimesFunc
	copy         []CopyFunc
	create       []CreateFunc
	createTemp   []CreateTempFunc
	glob         []GlobFunc
	lstat        []LstatFunc
	mkdir        []MkdirFunc
	mkdirAll     []MkdirAllFunc
	mkdirTemp    []MkdirTempFunc
	openFile     []OpenFileFunc
	readDir      []ReadDirFunc
	readDirNames []ReadDirNamesFunc
	readFile     []ReadFileFunc
	readLink     []ReadLinkFunc
	remove       []RemoveFunc
	removeAll    []RemoveAllFunc
	rename       []RenameFunc
	sub          []SubFunc
	symlink      []SymlinkFunc
	tempFile     []TempFileFunc
	writeFile    []WriteFileFunc
}

// NewFs creates a new factory [Fs].
func NewFs() *Fs {
	return &Fs{name: "testfs/factory"}
}

// Named sets the name of the filesystem and returns it.
func (f *Fs) Named(name string) *Fs {
	f.name = name
	return f
}

// Name returns the filesystem name.
func (f *Fs) Name() string {
	return f.name
}

// WithOpen appends Open functions to the factory queue.
func (f *Fs) WithOpen(open ...OpenFunc) *Fs {
	f.open = append(f.open, open...)
	return f
}

// SetOpen replaces the Open function queue.
func (f *Fs) SetOpen(open ...OpenFunc) *Fs {
	f.open = open
	return f
}

// Open implements [ihfs.FS] by consuming the next queued Open function.
func (f *Fs) Open(path string) (ihfs.File, error) {
	if len(f.open) == 0 {
		return nil, ErrNoMocks
	}

	open := f.open[0]
	f.open = f.open[1:]
	return open(path)
}

// WithStat appends Stat functions to the factory queue.
func (f *Fs) WithStat(stat ...StatFunc) *Fs {
	f.stat = append(f.stat, stat...)
	return f
}

// SetStat replaces the Stat function queue.
func (f *Fs) SetStat(stat ...StatFunc) *Fs {
	f.stat = stat
	return f
}

// Stat implements [ihfs.StatFS] by consuming the next queued Stat function.
func (f *Fs) Stat(path string) (ihfs.FileInfo, error) {
	if len(f.stat) == 0 {
		return nil, ErrNoMocks
	}

	stat := f.stat[0]
	f.stat = f.stat[1:]
	return stat(path)
}

func (f *Fs) WithChmod(chmod ...ChmodFunc) *Fs {
	f.chmod = append(f.chmod, chmod...)
	return f
}

func (f *Fs) SetChmod(chmod ...ChmodFunc) *Fs {
	f.chmod = chmod
	return f
}

func (f *Fs) Chmod(name string, mode ihfs.FileMode) error {
	if len(f.chmod) == 0 {
		return ErrNoMocks
	}

	chmod := f.chmod[0]
	f.chmod = f.chmod[1:]
	return chmod(name, mode)
}

func (f *Fs) WithChown(chown ...ChownFunc) *Fs {
	f.chown = append(f.chown, chown...)
	return f
}

func (f *Fs) SetChown(chown ...ChownFunc) *Fs {
	f.chown = chown
	return f
}

func (f *Fs) Chown(name string, uid, gid int) error {
	if len(f.chown) == 0 {
		return ErrNoMocks
	}

	chown := f.chown[0]
	f.chown = f.chown[1:]
	return chown(name, uid, gid)
}

func (f *Fs) WithChtimes(chtimes ...ChtimesFunc) *Fs {
	f.chtimes = append(f.chtimes, chtimes...)
	return f
}

func (f *Fs) SetChtimes(chtimes ...ChtimesFunc) *Fs {
	f.chtimes = chtimes
	return f
}

func (f *Fs) Chtimes(name string, atime, mtime time.Time) error {
	if len(f.chtimes) == 0 {
		return ErrNoMocks
	}

	chtimes := f.chtimes[0]
	f.chtimes = f.chtimes[1:]
	return chtimes(name, atime, mtime)
}

func (f *Fs) WithCopy(copy ...CopyFunc) *Fs {
	f.copy = append(f.copy, copy...)
	return f
}

func (f *Fs) SetCopy(copy ...CopyFunc) *Fs {
	f.copy = copy
	return f
}

func (f *Fs) Copy(dir string, fsys ihfs.FS) error {
	if len(f.copy) == 0 {
		return ErrNoMocks
	}

	copy := f.copy[0]
	f.copy = f.copy[1:]
	return copy(dir, fsys)
}

func (f *Fs) WithCreate(create ...CreateFunc) *Fs {
	f.create = append(f.create, create...)
	return f
}

func (f *Fs) SetCreate(create ...CreateFunc) *Fs {
	f.create = create
	return f
}

func (f *Fs) Create(name string) (ihfs.File, error) {
	if len(f.create) == 0 {
		return nil, ErrNoMocks
	}

	create := f.create[0]
	f.create = f.create[1:]
	return create(name)
}

func (f *Fs) WithCreateTemp(createTemp ...CreateTempFunc) *Fs {
	f.createTemp = append(f.createTemp, createTemp...)
	return f
}

func (f *Fs) SetCreateTemp(createTemp ...CreateTempFunc) *Fs {
	f.createTemp = createTemp
	return f
}

func (f *Fs) CreateTemp(dir, pattern string) (ihfs.File, error) {
	if len(f.createTemp) == 0 {
		return nil, ErrNoMocks
	}

	createTemp := f.createTemp[0]
	f.createTemp = f.createTemp[1:]
	return createTemp(dir, pattern)
}

func (f *Fs) WithGlob(glob ...GlobFunc) *Fs {
	f.glob = append(f.glob, glob...)
	return f
}

func (f *Fs) SetGlob(glob ...GlobFunc) *Fs {
	f.glob = glob
	return f
}

func (f *Fs) Glob(pattern string) ([]string, error) {
	if len(f.glob) == 0 {
		return nil, ErrNoMocks
	}

	glob := f.glob[0]
	f.glob = f.glob[1:]
	return glob(pattern)
}

func (f *Fs) WithLstat(lstat ...LstatFunc) *Fs {
	f.lstat = append(f.lstat, lstat...)
	return f
}

func (f *Fs) SetLstat(lstat ...LstatFunc) *Fs {
	f.lstat = lstat
	return f
}

func (f *Fs) Lstat(name string) (ihfs.FileInfo, error) {
	if len(f.lstat) == 0 {
		return nil, ErrNoMocks
	}

	lstat := f.lstat[0]
	f.lstat = f.lstat[1:]
	return lstat(name)
}

func (f *Fs) WithMkdir(mkdir ...MkdirFunc) *Fs {
	f.mkdir = append(f.mkdir, mkdir...)
	return f
}

func (f *Fs) SetMkdir(mkdir ...MkdirFunc) *Fs {
	f.mkdir = mkdir
	return f
}

func (f *Fs) Mkdir(name string, mode ihfs.FileMode) error {
	if len(f.mkdir) == 0 {
		return ErrNoMocks
	}

	mkdir := f.mkdir[0]
	f.mkdir = f.mkdir[1:]
	return mkdir(name, mode)
}

func (f *Fs) WithMkdirAll(mkdirAll ...MkdirAllFunc) *Fs {
	f.mkdirAll = append(f.mkdirAll, mkdirAll...)
	return f
}

func (f *Fs) SetMkdirAll(mkdirAll ...MkdirAllFunc) *Fs {
	f.mkdirAll = mkdirAll
	return f
}

func (f *Fs) MkdirAll(name string, mode ihfs.FileMode) error {
	if len(f.mkdirAll) == 0 {
		return ErrNoMocks
	}

	mkdirAll := f.mkdirAll[0]
	f.mkdirAll = f.mkdirAll[1:]
	return mkdirAll(name, mode)
}

func (f *Fs) WithMkdirTemp(mkdirTemp ...MkdirTempFunc) *Fs {
	f.mkdirTemp = append(f.mkdirTemp, mkdirTemp...)
	return f
}

func (f *Fs) SetMkdirTemp(mkdirTemp ...MkdirTempFunc) *Fs {
	f.mkdirTemp = mkdirTemp
	return f
}

func (f *Fs) MkdirTemp(dir, pattern string) (string, error) {
	if len(f.mkdirTemp) == 0 {
		return "", ErrNoMocks
	}

	mkdirTemp := f.mkdirTemp[0]
	f.mkdirTemp = f.mkdirTemp[1:]
	return mkdirTemp(dir, pattern)
}

func (f *Fs) WithOpenFile(openFile ...OpenFileFunc) *Fs {
	f.openFile = append(f.openFile, openFile...)
	return f
}

func (f *Fs) SetOpenFile(openFile ...OpenFileFunc) *Fs {
	f.openFile = openFile
	return f
}

func (f *Fs) OpenFile(name string, flag int, perm ihfs.FileMode) (ihfs.File, error) {
	if len(f.openFile) == 0 {
		return nil, ErrNoMocks
	}

	openFile := f.openFile[0]
	f.openFile = f.openFile[1:]
	return openFile(name, flag, perm)
}

func (f *Fs) WithReadDir(readDir ...ReadDirFunc) *Fs {
	f.readDir = append(f.readDir, readDir...)
	return f
}

func (f *Fs) SetReadDir(readDir ...ReadDirFunc) *Fs {
	f.readDir = readDir
	return f
}

func (f *Fs) ReadDir(name string) ([]ihfs.DirEntry, error) {
	if len(f.readDir) == 0 {
		return nil, ErrNoMocks
	}

	readDir := f.readDir[0]
	f.readDir = f.readDir[1:]
	return readDir(name)
}

func (f *Fs) WithReadDirNames(readDirNames ...ReadDirNamesFunc) *Fs {
	f.readDirNames = append(f.readDirNames, readDirNames...)
	return f
}

func (f *Fs) SetReadDirNames(readDirNames ...ReadDirNamesFunc) *Fs {
	f.readDirNames = readDirNames
	return f
}

func (f *Fs) ReadDirNames(name string) ([]string, error) {
	if len(f.readDirNames) == 0 {
		return nil, ErrNoMocks
	}

	readDirNames := f.readDirNames[0]
	f.readDirNames = f.readDirNames[1:]
	return readDirNames(name)
}

func (f *Fs) WithReadFile(readFile ...ReadFileFunc) *Fs {
	f.readFile = append(f.readFile, readFile...)
	return f
}

func (f *Fs) SetReadFile(readFile ...ReadFileFunc) *Fs {
	f.readFile = readFile
	return f
}

func (f *Fs) ReadFile(name string) ([]byte, error) {
	if len(f.readFile) == 0 {
		return nil, ErrNoMocks
	}

	readFile := f.readFile[0]
	f.readFile = f.readFile[1:]
	return readFile(name)
}

func (f *Fs) WithReadLink(readLink ...ReadLinkFunc) *Fs {
	f.readLink = append(f.readLink, readLink...)
	return f
}

func (f *Fs) SetReadLink(readLink ...ReadLinkFunc) *Fs {
	f.readLink = readLink
	return f
}

func (f *Fs) ReadLink(name string) (string, error) {
	if len(f.readLink) == 0 {
		return "", ErrNoMocks
	}

	readLink := f.readLink[0]
	f.readLink = f.readLink[1:]
	return readLink(name)
}

func (f *Fs) WithRemove(remove ...RemoveFunc) *Fs {
	f.remove = append(f.remove, remove...)
	return f
}

func (f *Fs) SetRemove(remove ...RemoveFunc) *Fs {
	f.remove = remove
	return f
}

func (f *Fs) Remove(name string) error {
	if len(f.remove) == 0 {
		return ErrNoMocks
	}

	remove := f.remove[0]
	f.remove = f.remove[1:]
	return remove(name)
}

func (f *Fs) WithRemoveAll(removeAll ...RemoveAllFunc) *Fs {
	f.removeAll = append(f.removeAll, removeAll...)
	return f
}

func (f *Fs) SetRemoveAll(removeAll ...RemoveAllFunc) *Fs {
	f.removeAll = removeAll
	return f
}

func (f *Fs) RemoveAll(name string) error {
	if len(f.removeAll) == 0 {
		return ErrNoMocks
	}

	removeAll := f.removeAll[0]
	f.removeAll = f.removeAll[1:]
	return removeAll(name)
}

func (f *Fs) WithRename(rename ...RenameFunc) *Fs {
	f.rename = append(f.rename, rename...)
	return f
}

func (f *Fs) SetRename(rename ...RenameFunc) *Fs {
	f.rename = rename
	return f
}

func (f *Fs) Rename(oldpath, newpath string) error {
	if len(f.rename) == 0 {
		return ErrNoMocks
	}

	rename := f.rename[0]
	f.rename = f.rename[1:]
	return rename(oldpath, newpath)
}

func (f *Fs) WithSub(sub ...SubFunc) *Fs {
	f.sub = append(f.sub, sub...)
	return f
}

func (f *Fs) SetSub(sub ...SubFunc) *Fs {
	f.sub = sub
	return f
}

func (f *Fs) Sub(dir string) (ihfs.FS, error) {
	if len(f.sub) == 0 {
		return nil, ErrNoMocks
	}

	sub := f.sub[0]
	f.sub = f.sub[1:]
	return sub(dir)
}

func (f *Fs) WithSymlink(symlink ...SymlinkFunc) *Fs {
	f.symlink = append(f.symlink, symlink...)
	return f
}

func (f *Fs) SetSymlink(symlink ...SymlinkFunc) *Fs {
	f.symlink = symlink
	return f
}

func (f *Fs) Symlink(oldname, newname string) error {
	if len(f.symlink) == 0 {
		return ErrNoMocks
	}

	symlink := f.symlink[0]
	f.symlink = f.symlink[1:]
	return symlink(oldname, newname)
}

func (f *Fs) WithTempFile(tempFile ...TempFileFunc) *Fs {
	f.tempFile = append(f.tempFile, tempFile...)
	return f
}

func (f *Fs) SetTempFile(tempFile ...TempFileFunc) *Fs {
	f.tempFile = tempFile
	return f
}

func (f *Fs) TempFile(dir, pattern string) (string, error) {
	if len(f.tempFile) == 0 {
		return "", ErrNoMocks
	}

	tempFile := f.tempFile[0]
	f.tempFile = f.tempFile[1:]
	return tempFile(dir, pattern)
}

func (f *Fs) WithWriteFile(writeFile ...WriteFileFunc) *Fs {
	f.writeFile = append(f.writeFile, writeFile...)
	return f
}

func (f *Fs) SetWriteFile(writeFile ...WriteFileFunc) *Fs {
	f.writeFile = writeFile
	return f
}

func (f *Fs) WriteFile(name string, data []byte, perm ihfs.FileMode) error {
	if len(f.writeFile) == 0 {
		return ErrNoMocks
	}

	writeFile := f.writeFile[0]
	f.writeFile = f.writeFile[1:]
	return writeFile(name, data, perm)
}

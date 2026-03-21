package memfs

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/unstoppablemango/ihfs"
)

var separator = string(filepath.Separator)

// Fs represents an in-memory filesystem.
type Fs struct {
	mu      sync.RWMutex
	data    map[string]*FileData
	init    sync.Once
	tempSeq atomic.Uint64
}

// New creates a new in-memory filesystem.
func New() *Fs {
	return &Fs{}
}

func (f *Fs) getData() map[string]*FileData {
	f.init.Do(func() {
		f.data = make(map[string]*FileData)
		// Root should always exist
		root := CreateDir(separator)
		f.data[separator] = root
	})
	return f.data
}

// Open implements ihfs.FS.
func (f *Fs) Open(name string) (ihfs.File, error) {
	// Store original name for error messages
	origName := name

	// Validate path before normalization
	if !fs.ValidPath(name) {
		return nil, perror("open", origName, ihfs.ErrInvalid)
	}

	// Normalize after validation for internal use
	name = normalizePath(name)

	f.mu.RLock()
	file, ok := f.getData()[name]
	f.mu.RUnlock()

	if !ok {
		return nil, perror("open", origName, ihfs.ErrNotExist)
	}

	return NewReadOnlyFile(file), nil
}

// Create implements ihfs.CreateFS.
func (f *Fs) Create(name string) (ihfs.File, error) {
	name = normalizePath(name)

	f.mu.Lock()
	defer f.mu.Unlock()

	file := CreateFile(name)
	f.getData()[name] = file

	if err := f.registerWithParent(file); err != nil {
		delete(f.getData(), name)
		return nil, perror("create", name, err)
	}

	return NewFile(file), nil
}

// Mkdir implements ihfs.MkdirFS.
func (f *Fs) Mkdir(name string, perm os.FileMode) error {
	name = normalizePath(name)

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.getData()[name]; exists {
		return perror("mkdir", name, ihfs.ErrExist)
	}

	dir := CreateDir(name)
	dir.mode = os.ModeDir | perm
	f.getData()[name] = dir

	if err := f.registerWithParent(dir); err != nil {
		return perror("mkdir", name, err)
	}

	return nil
}

// MkdirAll implements ihfs.MkdirAllFS.
func (f *Fs) MkdirAll(name string, perm os.FileMode) error {
	name = normalizePath(name)

	f.mu.Lock()
	defer f.mu.Unlock()

	// Check if it already exists
	if file, exists := f.getData()[name]; exists {
		if !file.isDir {
			return perror("mkdirall", name, ihfs.ErrExist)
		}
		return nil
	}

	// Create all parent directories
	parts := strings.Split(strings.Trim(name, separator), separator)
	current := separator

	for _, part := range parts {
		current = filepath.Join(current, part)
		if _, exists := f.getData()[current]; !exists {
			dir := CreateDir(current)
			dir.mode = os.ModeDir | perm
			f.getData()[current] = dir

			if err := f.registerWithParent(dir); err != nil {
				return perror("mkdirall", name, err)
			}
		}
	}

	return nil
}

// Remove implements ihfs.RemoveFS.
func (f *Fs) Remove(name string) error {
	name = normalizePath(name)

	f.mu.Lock()
	defer f.mu.Unlock()

	file, ok := f.getData()[name]
	if !ok {
		return perror("remove", name, ihfs.ErrNotExist)
	}

	// Check if directory is empty
	if file.isDir && file.dir != nil {
		file.dir.Lock()
		isEmpty := len(file.dir.children) == 0
		file.dir.Unlock()

		if !isEmpty {
			return perror("remove", name, ihfs.ErrInvalid)
		}
	}

	f.unregisterWithParent(name)

	delete(f.getData(), name)
	return nil
}

// RemoveAll implements ihfs.RemoveAllFS.
func (f *Fs) RemoveAll(name string) error {
	name = normalizePath(name)

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.getData()[name]; !ok {
		return nil // RemoveAll doesn't error if path doesn't exist
	}

	// Find all descendants
	descendants := f.findDescendants(name)

	// Remove descendants first (depth-first)
	for i := len(descendants) - 1; i >= 0; i-- {
		delete(f.getData(), descendants[i].name)
	}

	// Unregister with parent
	f.unregisterWithParent(name)

	// Remove the target
	delete(f.getData(), name)
	return nil
}

// Rename implements ihfs.RenameFS.
func (f *Fs) Rename(oldName, newName string) error {
	oldName = normalizePath(oldName)
	newName = normalizePath(newName)

	f.mu.Lock()
	defer f.mu.Unlock()

	file, ok := f.getData()[oldName]
	if !ok {
		return perror("rename", oldName, ihfs.ErrNotExist)
	}
	if _, exists := f.getData()[newName]; exists {
		return perror("rename", newName, ihfs.ErrExist)
	}

	// Validate new parent directory exists and is a directory BEFORE making any changes
	// This prevents leaving the filesystem in an inconsistent state if validation fails
	newParentPath := filepath.Dir(newName)
	if newParentPath != "/" {
		newParent, exists := f.getData()[newParentPath]
		if !exists {
			return perror("rename", newName, ihfs.ErrNotExist)
		}
		if !newParent.isDir {
			return perror("rename", newName, ihfs.ErrInvalid)
		}
	}

	// Now that validation is complete, we can safely make changes
	f.unregisterWithParent(oldName)

	file.Lock()
	file.name = newName
	file.Unlock()

	delete(f.getData(), oldName)
	f.getData()[newName] = file

	return f.registerWithParent(file)
}

// Stat implements ihfs.StatFS.
func (f *Fs) Stat(name string) (ihfs.FileInfo, error) {
	name = normalizePath(name)

	f.mu.RLock()
	file, ok := f.getData()[name]
	f.mu.RUnlock()

	if !ok {
		return nil, perror("stat", name, ihfs.ErrNotExist)
	}

	return &FileInfo{data: file}, nil
}

// Chmod implements ihfs.ChmodFS.
func (f *Fs) Chmod(name string, mode os.FileMode) error {
	name = normalizePath(name)

	f.mu.RLock()
	file, ok := f.getData()[name]
	f.mu.RUnlock()

	if !ok {
		return perror("chmod", name, ihfs.ErrNotExist)
	}

	file.Lock()
	file.mode = mode
	file.Unlock()

	return nil
}

// Chown implements ihfs.ChownFS.
func (f *Fs) Chown(name string, uid, gid int) error {
	name = normalizePath(name)

	f.mu.RLock()
	file, ok := f.getData()[name]
	f.mu.RUnlock()

	if !ok {
		return perror("chown", name, ihfs.ErrNotExist)
	}

	file.Lock()
	file.uid = uid
	file.gid = gid
	file.Unlock()

	return nil
}

// Chtimes implements ihfs.ChtimesFS.
func (f *Fs) Chtimes(name string, _, mtime time.Time) error {
	name = normalizePath(name)

	f.mu.RLock()
	file, ok := f.getData()[name]
	f.mu.RUnlock()

	if !ok {
		return perror("chtimes", name, ihfs.ErrNotExist)
	}

	file.Lock()
	file.modTime = mtime
	file.Unlock()

	return nil
}

// OpenFile implements ihfs.OpenFileFS.
func (f *Fs) OpenFile(name string, flag int, perm os.FileMode) (ihfs.File, error) {
	name = normalizePath(name)

	f.mu.Lock()
	defer f.mu.Unlock()

	file, exists := f.getData()[name]

	if !exists {
		if flag&os.O_CREATE == 0 {
			return nil, perror("open", name, ihfs.ErrNotExist)
		}

		file = CreateFile(name)
		file.mode = perm
		f.getData()[name] = file

		if err := f.registerWithParent(file); err != nil {
			delete(f.getData(), name)
			return nil, perror("open", name, err)
		}
	} else if flag&os.O_EXCL != 0 {
		return nil, perror("open", name, ihfs.ErrExist)
	}

	if flag&os.O_TRUNC != 0 && !file.isDir {
		file.Lock()
		file.content = []byte{}
		file.Unlock()
	}

	handle := NewFile(file)
	if flag&os.O_APPEND != 0 {
		file.Lock()
		handle.at = int64(len(file.content))
		file.Unlock()
	}

	if flag&(os.O_WRONLY|os.O_RDWR) == 0 {
		handle.readOnly = true
	}

	return handle, nil
}

// WriteFile implements ihfs.WriteFileFS.
func (f *Fs) WriteFile(name string, data []byte, perm ihfs.FileMode) error {
	file, err := f.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err = file.(ihfs.Writer).Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func (f *Fs) registerWithParent(file *FileData) error {
	parent := f.findParent(file)
	if parent == nil {
		return ihfs.ErrNotExist
	}
	if !parent.isDir {
		return ihfs.ErrInvalid
	}

	parent.dir.Lock()
	defer parent.dir.Unlock()

	baseName := filepath.Base(file.name)
	parent.dir.children[baseName] = file

	return nil
}

func (f *Fs) unregisterWithParent(name string) {
	file := f.getData()[name]
	parent := f.findParent(file)
	if parent == nil {
		// Root has no parent
		return
	}

	parent.dir.Lock()
	defer parent.dir.Unlock()

	baseName := filepath.Base(name)
	delete(parent.dir.children, baseName)
}

func (f *Fs) findParent(file *FileData) *FileData {
	parentPath := filepath.Dir(file.name)
	if parentPath == file.name {
		// We're at root
		return nil
	}

	return f.getData()[parentPath]
}

func (f *Fs) findDescendants(name string) []*FileData {
	var descendants []*FileData
	prefix := name + separator

	for path, file := range f.getData() {
		if strings.HasPrefix(path, prefix) {
			descendants = append(descendants, file)
		}
	}

	return descendants
}

// ReadDir implements ihfs.ReadDirFS.
func (f *Fs) ReadDir(name string) ([]ihfs.DirEntry, error) {
	name = normalizePath(name)

	f.mu.RLock()
	file, ok := f.getData()[name]
	f.mu.RUnlock()

	if !ok {
		return nil, perror("readdir", name, ihfs.ErrNotExist)
	}
	if !file.isDir {
		return nil, perror("readdir", name, ihfs.ErrInvalid)
	}

	file.dir.Lock()
	defer file.dir.Unlock()

	entries := make([]ihfs.DirEntry, 0, len(file.dir.children))
	for _, child := range file.dir.children {
		entries = append(entries, &FileInfo{data: child})
	}

	sortDirEntries(entries)
	return entries, nil
}

// WriteFile implements ihfs.WriteFileFS.
func (f *Fs) WriteFile(name string, data []byte, perm ihfs.FileMode) error {
	normalName := normalizePath(name)

	f.mu.Lock()
	defer f.mu.Unlock()

	file, exists := f.getData()[normalName]
	if !exists {
		file = CreateFile(normalName)
		file.mode = perm
		f.getData()[normalName] = file

		if err := f.registerWithParent(file); err != nil {
			delete(f.getData(), normalName)
			return perror("writefile", name, err)
		}
	}

	file.Lock()
	file.content = make([]byte, len(data))
	copy(file.content, data)
	file.modTime = time.Now()
	file.Unlock()

	return nil
}

// Copy implements ihfs.CopyFS.
func (f *Fs) Copy(dir string, src ihfs.FS) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		var destPath string
		if path == "." {
			destPath = normalizePath(dir)
		} else {
			destPath = filepath.Join(normalizePath(dir), path)
		}

		if d.IsDir() {
			if path == "." {
				f.mu.RLock()
				_, exists := f.getData()[destPath]
				f.mu.RUnlock()
				if !exists {
					return f.Mkdir(destPath, 0755)
				}
				return nil
			}
			return f.Mkdir(destPath, 0755)
		}

		f.mu.RLock()
		_, exists := f.getData()[destPath]
		f.mu.RUnlock()
		if exists {
			return perror("copy", destPath, ihfs.ErrExist)
		}

		srcFile, err := src.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = srcFile.Close() }()

		info, err := d.Info()
		if err != nil {
			return err
		}

		data, err := io.ReadAll(srcFile)
		if err != nil {
			return err
		}

		return f.WriteFile(destPath, data, info.Mode())
	})
}

// nextTempName generates a unique name for a temporary file or directory.
func (f *Fs) nextTempName(dir, pattern, op string) (string, error) {
	prefix, suffix, _ := strings.Cut(pattern, "*")
	normalDir := normalizePath(dir)

	f.mu.RLock()
	dirData, ok := f.getData()[normalDir]
	f.mu.RUnlock()

	if !ok {
		return "", perror(op, dir, ihfs.ErrNotExist)
	}
	if !dirData.isDir {
		return "", perror(op, dir, ihfs.ErrInvalid)
	}

	seq := strconv.FormatUint(f.tempSeq.Add(1), 10)
	return filepath.Join(normalDir, prefix+seq+suffix), nil
}

// CreateTemp implements ihfs.CreateTempFS.
func (f *Fs) CreateTemp(dir, pattern string) (ihfs.File, error) {
	name, err := f.nextTempName(dir, pattern, "createtemp")
	if err != nil {
		return nil, err
	}
	return f.Create(name)
}

// MkdirTemp implements ihfs.MkdirTempFS.
func (f *Fs) MkdirTemp(dir, pattern string) (string, error) {
	name, err := f.nextTempName(dir, pattern, "mkdirtemp")
	if err != nil {
		return "", err
	}
	if err := f.Mkdir(name, 0700); err != nil {
		return "", err
	}
	return toIOFSPath(name), nil
}

// TempFile implements ihfs.TempFileFS.
func (f *Fs) TempFile(dir, pattern string) (string, error) {
	name, err := f.nextTempName(dir, pattern, "tempfile")
	if err != nil {
		return "", err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	file := CreateFile(name)
	f.getData()[name] = file

	if err := f.registerWithParent(file); err != nil {
		delete(f.getData(), name)
		return "", perror("tempfile", name, err)
	}

	return toIOFSPath(name), nil
}

// Symlink implements ihfs.SymlinkFS.
func (f *Fs) Symlink(oldname, newname string) error {
	normalNew := normalizePath(newname)

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.getData()[normalNew]; exists {
		return perror("symlink", newname, ihfs.ErrExist)
	}

	link := &FileData{
		name:       normalNew,
		isSymlink:  true,
		linkTarget: oldname,
		mode:       os.ModeSymlink | 0777,
		modTime:    time.Now(),
	}
	f.getData()[normalNew] = link

	if err := f.registerWithParent(link); err != nil {
		delete(f.getData(), normalNew)
		return err
	}
	return nil
}

// ReadLink implements ihfs.ReadLinkFS.
func (f *Fs) ReadLink(name string) (string, error) {
	name = normalizePath(name)

	f.mu.RLock()
	file, ok := f.getData()[name]
	f.mu.RUnlock()

	if !ok {
		return "", perror("readlink", name, ihfs.ErrNotExist)
	}

	file.Lock()
	defer file.Unlock()

	if !file.isSymlink {
		return "", perror("readlink", name, ihfs.ErrInvalid)
	}

	return file.linkTarget, nil
}

// Lstat implements ihfs.ReadLinkFS. It returns info about the named file
// without following symbolic links.
func (f *Fs) Lstat(name string) (ihfs.FileInfo, error) {
	name = normalizePath(name)

	f.mu.RLock()
	file, ok := f.getData()[name]
	f.mu.RUnlock()

	if !ok {
		return nil, perror("lstat", name, ihfs.ErrNotExist)
	}

	return &FileInfo{data: file}, nil
}

func toIOFSPath(internalPath string) string {
	return strings.TrimPrefix(internalPath, separator)
}

func normalizePath(path string) string {
	// Convert io/fs style paths to internal absolute paths
	if path == "." {
		return separator
	}
	if path == "" {
		// Treat empty paths as referring to the filesystem root
		return separator
	}

	// Prepend "/" for relative paths (io/fs style)
	if !strings.HasPrefix(path, separator) {
		path = separator + path
	}

	return filepath.Clean(path)
}

func perror(op, path string, err error) error {
	return &ihfs.PathError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

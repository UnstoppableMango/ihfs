package memfs

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/unstoppablemango/ihfs"
)

// Fs represents an in-memory filesystem.
type Fs struct {
	mu   sync.RWMutex
	data map[string]*FileData
	init sync.Once
}

// New creates a new in-memory filesystem.
func New() *Fs {
	return &Fs{}
}

func (f *Fs) getData() map[string]*FileData {
	f.init.Do(func() {
		f.data = make(map[string]*FileData)
		// Root should always exist
		root := CreateDir(string(filepath.Separator))
		f.data[string(filepath.Separator)] = root
	})
	return f.data
}

// Open implements ihfs.FS.
func (f *Fs) Open(name string) (ihfs.File, error) {
	name = normalizePath(name)

	f.mu.RLock()
	file, ok := f.getData()[name]
	f.mu.RUnlock()

	if !ok {
		return nil, perror("open", name, ihfs.ErrNotExist)
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
	parts := strings.Split(strings.Trim(name, string(filepath.Separator)), string(filepath.Separator))
	current := string(filepath.Separator)

	for _, part := range parts {
		if part == "" {
			continue
		}

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

	if err := f.registerWithParent(file); err != nil {
		return perror("rename", newName, err)
	}

	return nil
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
func (f *Fs) Chtimes(name string, atime, mtime time.Time) error {
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
	prefix := name + string(filepath.Separator)

	for path, file := range f.getData() {
		if strings.HasPrefix(path, prefix) {
			descendants = append(descendants, file)
		}
	}

	return descendants
}

func normalizePath(path string) string {
	if path == "" {
		return string(filepath.Separator)
	}

	path = filepath.Clean(path)
	if !strings.HasPrefix(path, string(filepath.Separator)) {
		path = string(filepath.Separator) + path
	}

	return path
}

func perror(op, path string, err error) error {
	return &ihfs.PathError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

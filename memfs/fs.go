package memfs

import (
	"io/fs"
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

func (m *Fs) getData() map[string]*FileData {
	m.init.Do(func() {
		m.data = make(map[string]*FileData)
		// Root should always exist
		root := CreateDir(string(filepath.Separator))
		m.data[string(filepath.Separator)] = root
	})
	return m.data
}

// Open implements ihfs.FS.
func (m *Fs) Open(name string) (ihfs.File, error) {
	name = normalizePath(name)

	m.mu.RLock()
	file, ok := m.getData()[name]
	m.mu.RUnlock()

	if !ok {
		return nil, &ihfs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	return NewReadOnlyFile(file), nil
}

// Create implements ihfs.CreateFS.
func (m *Fs) Create(name string) (ihfs.File, error) {
	name = normalizePath(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	file := CreateFile(name)
	m.getData()[name] = file

	if err := m.registerWithParent(file); err != nil {
		delete(m.getData(), name)
		return nil, err
	}

	return NewFile(file), nil
}

// Mkdir implements ihfs.MkdirFS.
func (m *Fs) Mkdir(name string, perm os.FileMode) error {
	name = normalizePath(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.getData()[name]; exists {
		return &ihfs.PathError{Op: "mkdir", Path: name, Err: fs.ErrExist}
	}

	dir := CreateDir(name)
	dir.mode = os.ModeDir | perm
	m.getData()[name] = dir

	return m.registerWithParent(dir)
}

// MkdirAll implements ihfs.MkdirAllFS.
func (m *Fs) MkdirAll(name string, perm os.FileMode) error {
	name = normalizePath(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if it already exists
	if file, exists := m.getData()[name]; exists {
		if !file.isDir {
			return &ihfs.PathError{Op: "mkdir", Path: name, Err: fs.ErrExist}
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

		if _, exists := m.getData()[current]; !exists {
			dir := CreateDir(current)
			dir.mode = os.ModeDir | perm
			m.getData()[current] = dir

			if err := m.registerWithParent(dir); err != nil {
				return err
			}
		}
	}

	return nil
}

// Remove implements ihfs.RemoveFS.
func (m *Fs) Remove(name string) error {
	name = normalizePath(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	file, ok := m.getData()[name]
	if !ok {
		return &ihfs.PathError{Op: "remove", Path: name, Err: fs.ErrNotExist}
	}

	// Check if directory is empty
	if file.isDir && file.dir != nil {
		file.dir.Lock()
		isEmpty := len(file.dir.children) == 0
		file.dir.Unlock()

		if !isEmpty {
			return &ihfs.PathError{Op: "remove", Path: name, Err: os.ErrInvalid}
		}
	}

	m.unregisterWithParent(name)

	delete(m.getData(), name)
	return nil
}

// RemoveAll implements ihfs.RemoveAllFS.
func (m *Fs) RemoveAll(name string) error {
	name = normalizePath(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.getData()[name]; !ok {
		return nil // RemoveAll doesn't error if path doesn't exist
	}

	// Find all descendants
	descendants := m.findDescendants(name)

	// Remove descendants first (depth-first)
	for i := len(descendants) - 1; i >= 0; i-- {
		delete(m.getData(), descendants[i].name)
	}

	// Unregister with parent
	m.unregisterWithParent(name)

	// Remove the target
	delete(m.getData(), name)
	return nil
}

// Rename implements ihfs.RenameFS.
func (m *Fs) Rename(oldName, newName string) error {
	oldName = normalizePath(oldName)
	newName = normalizePath(newName)

	m.mu.Lock()
	defer m.mu.Unlock()

	file, ok := m.getData()[oldName]
	if !ok {
		return &ihfs.PathError{Op: "rename", Path: oldName, Err: fs.ErrNotExist}
	}

	if _, exists := m.getData()[newName]; exists {
		return &ihfs.PathError{Op: "rename", Path: newName, Err: fs.ErrExist}
	}

	// Unregister from old parent
	m.unregisterWithParent(oldName)

	// Update name
	file.Lock()
	file.name = newName
	file.Unlock()

	// Update map
	delete(m.getData(), oldName)
	m.getData()[newName] = file

	// Register with new parent
	return m.registerWithParent(file)
}

// Stat implements ihfs.StatFS.
func (m *Fs) Stat(name string) (ihfs.FileInfo, error) {
	name = normalizePath(name)

	m.mu.RLock()
	file, ok := m.getData()[name]
	m.mu.RUnlock()

	if !ok {
		return nil, &ihfs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
	}

	return &FileInfo{data: file}, nil
}

// Chmod implements ihfs.ChmodFS.
func (m *Fs) Chmod(name string, mode os.FileMode) error {
	name = normalizePath(name)

	m.mu.RLock()
	file, ok := m.getData()[name]
	m.mu.RUnlock()

	if !ok {
		return &ihfs.PathError{Op: "chmod", Path: name, Err: fs.ErrNotExist}
	}

	file.Lock()
	file.mode = mode
	file.Unlock()

	return nil
}

// Chown implements ihfs.ChownFS.
func (m *Fs) Chown(name string, uid, gid int) error {
	name = normalizePath(name)

	m.mu.RLock()
	file, ok := m.getData()[name]
	m.mu.RUnlock()

	if !ok {
		return &ihfs.PathError{Op: "chown", Path: name, Err: fs.ErrNotExist}
	}

	file.Lock()
	file.uid = uid
	file.gid = gid
	file.Unlock()

	return nil
}

// Chtimes implements ihfs.ChtimesFS.
func (m *Fs) Chtimes(name string, atime, mtime time.Time) error {
	name = normalizePath(name)

	m.mu.RLock()
	file, ok := m.getData()[name]
	m.mu.RUnlock()

	if !ok {
		return &ihfs.PathError{Op: "chtimes", Path: name, Err: fs.ErrNotExist}
	}

	file.Lock()
	file.modTime = mtime
	file.Unlock()

	return nil
}

// OpenFile implements ihfs.OpenFileFS.
func (m *Fs) OpenFile(name string, flag int, perm os.FileMode) (ihfs.File, error) {
	name = normalizePath(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	file, exists := m.getData()[name]

	if !exists {
		if flag&os.O_CREATE == 0 {
			return nil, &ihfs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
		}

		file = CreateFile(name)
		file.mode = perm
		m.getData()[name] = file

		if err := m.registerWithParent(file); err != nil {
			delete(m.getData(), name)
			return nil, err
		}
	} else if flag&os.O_EXCL != 0 {
		return nil, &ihfs.PathError{Op: "open", Path: name, Err: fs.ErrExist}
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

func (m *Fs) registerWithParent(file *FileData) error {
	parent := m.findParent(file)
	if parent == nil {
		return &ihfs.PathError{Op: "register", Path: file.name, Err: fs.ErrNotExist}
	}

	if !parent.isDir {
		return &ihfs.PathError{Op: "register", Path: file.name, Err: os.ErrInvalid}
	}

	parent.dir.Lock()
	defer parent.dir.Unlock()

	baseName := filepath.Base(file.name)
	parent.dir.children[baseName] = file

	return nil
}

func (m *Fs) unregisterWithParent(name string) {
	file := m.getData()[name]
	parent := m.findParent(file)
	if parent == nil {
		// Root has no parent
		return
	}

	parent.dir.Lock()
	defer parent.dir.Unlock()

	baseName := filepath.Base(name)
	delete(parent.dir.children, baseName)
}

func (m *Fs) findParent(file *FileData) *FileData {
	parentPath := filepath.Dir(file.name)
	if parentPath == file.name {
		// We're at root
		return nil
	}

	return m.getData()[parentPath]
}

func (m *Fs) findDescendants(name string) []*FileData {
	var descendants []*FileData
	prefix := name + string(filepath.Separator)

	for path, file := range m.getData() {
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

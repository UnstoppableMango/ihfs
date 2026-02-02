package op

import "io/fs"

// Open represents an operation to open a file.
type Open struct {
	Name string
}

// Subject implements [ihfs.Operation].
func (o Open) Subject() string {
	return o.Name
}

// Glob represents an operation to match files using a glob pattern.
type Glob struct {
	Pattern string
}

// Subject implements [ihfs.Operation].
func (g Glob) Subject() string {
	return g.Pattern
}

// Lstat represents an operation to get file information about symbolic links.
type Lstat struct {
	Name string
}

// Subject implements [ihfs.Operation].
func (l Lstat) Subject() string {
	return l.Name
}

// ReadDir represents an operation to read the contents of a directory.
type ReadDir struct {
	Name string
}

// Subject implements [ihfs.Operation].
func (r ReadDir) Subject() string {
	return r.Name
}

// ReadFile represents an operation to read a file.
type ReadFile struct {
	Name string
}

// Subject implements [ihfs.Operation].
func (r ReadFile) Subject() string {
	return r.Name
}

// ReadLink represents an operation to read a symbolic link.
type ReadLink struct {
	Name string
}

// Subject implements [ihfs.Operation].
func (r ReadLink) Subject() string {
	return r.Name
}

// Stat represents an operation to get file information.
type Stat struct {
	Name string
}

// Subject implements [ihfs.Operation].
func (s Stat) Subject() string {
	return s.Name
}

// WriteFile represents an operation to write to a file.
type WriteFile struct {
	Name string
	Data []byte
	Perm fs.FileMode
}

// Subject implements [ihfs.Operation].
func (w WriteFile) Subject() string {
	return w.Name
}

// Remove represents an operation to remove a file.
type Remove struct {
	Name string
}

// Subject implements [ihfs.Operation].
func (r Remove) Subject() string {
	return r.Name
}

// RemoveAll represents an operation to remove a directory and its contents.
type RemoveAll struct {
	Name string
}

// Subject implements [ihfs.Operation].
func (r RemoveAll) Subject() string {
	return r.Name
}

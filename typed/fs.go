package typed

import (
	"io"
	"io/fs"
)

type DirEntry[T fs.FileInfo] interface {
	Name() string
	IsDir() bool
	Type() fs.FileMode
	Info() (T, error)
}

type File[T fs.FileInfo] interface {
	io.ReadCloser
	Stat() (T, error)
}

type Directory[T fs.FileInfo] interface {
	io.Closer
	ReadDir(n int) ([]DirEntry[T], error)
}

type FS[T File[V], V fs.FileInfo] interface {
	Open(name string) (T, error)
}

package typed

import (
	"io"

	"github.com/unstoppablemango/ihfs"
)

type DirEntry[T ihfs.FileInfo] interface {
	Name() string
	IsDir() bool
	Type() ihfs.FileMode
	Info() (T, error)
}

type File[T ihfs.FileInfo] interface {
	io.ReadCloser

	Stat() (T, error)
}

type Directory[T ihfs.FileInfo] interface {
	io.Closer

	ReadDir(n int) ([]DirEntry[T], error)
}

type FS[F File[FI], FI ihfs.FileInfo] interface {
	Open(name string) (F, error)
}

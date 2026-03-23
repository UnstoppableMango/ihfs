package filter

import (
	"io/fs"

	"github.com/unstoppablemango/ihfs/op"
)

type Filter[T fs.FS] interface {
	Matches(T, op.Operation) error
}

type (
	Func[T fs.FS]      func(Fs[T], op.Operation) error
	Predicate[T fs.FS] func(Fs[T], op.Operation) bool
)

type Fs[T fs.FS] struct {
	fs  T
	ops operations[T]
}

func (f Fs[T]) Base() T      { return f.fs }
func (f Fs[T]) Name() string { return "filter" }

// Stat implements [StatFS].
func (f Fs[T]) Stat(name string) (fs.FileInfo, error) {
	if err := f.ops.Stat(name); err != nil {
		return nil, err
	}
	return fs.Stat(f.fs, name)
}

// Open implements [FS].
func (f Fs[T]) Open(name string) (fs.File, error) {
	if err := f.ops.Open(name); err != nil {
		return nil, err
	}
	return f.fs.Open(name)
}

type operations[T fs.FS] struct {
	fs     T
	filter Filter[T]
}

func (o operations[T]) Open(name string) error {
	return o.Matches(op.Open{Name: name})
}

func (o operations[T]) Stat(name string) error {
	return o.Matches(op.Stat{Name: name})
}

func (o operations[T]) Matches(op op.Operation) error {
	return o.filter.Matches(o.fs, op)
}

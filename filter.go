package ihfs

import (
	"io/fs"

	"github.com/unstoppablemango/ihfs/op"
)

type (
	// Operation is an alias for [op.Operation].
	Operation = op.Operation
)

type Filter[T fs.FS] interface {
	Matches(T, Operation) error
}

// FilterFS is a file system that applies filter functions to operations
// before delegating them to the underlying file system.
type FilterFS[T fs.FS] struct {
	fs     T
	filter Filter[T]
}

// FilterWith creates a new [FilterFS] that wraps the given file system with the provided filter functions.
func FilterWith[T fs.FS](fsys T, filter ...Filter[T]) *FilterFS[T] {
	return &FilterFS[T]{
		fs:     fsys,
		filter: filters[T](filter),
	}
}

// Base implements [Decorator].
func (f *FilterFS[T]) Base() FS {
	return f.fs
}

// Name returns the name of the filter filesystem.
func (f *FilterFS[T]) Name() string {
	return "filter"
}

// Stat implements [StatFS].
func (f *FilterFS[T]) Stat(name string) (FileInfo, error) {
	op := op.Stat{Name: name}
	if err := f.filter.Matches(f.fs, op); err != nil {
		return nil, err
	}
	return Stat(f.fs, name)
}

// Open implements [FS].
func (f *FilterFS[T]) Open(name string) (File, error) {
	op := op.Open{Name: name}
	if err := f.filter.Matches(f.fs, op); err != nil {
		return nil, err
	}
	return f.fs.Open(name)
}

// Where creates a new [FilterFS] that applies the given predicates to
// operations before delegating them to the underlying file system.
func Where(fsys FS, predicates ...Predicate) *FilterFS {
	var filters []FilterFunc
	for _, p := range predicates {
		filters = append(filters, p.Filter)
	}
	return FilterWith(fsys, filters...)
}

type filters[T fs.FS] []Filter[T]

func (fs filters[T]) Matches(f T, op Operation) error {
	for _, filter := range fs {
		if err := filter.Matches(f, op); err != nil {
			return err
		}
	}
	return nil
}

func flat[T fs.FS](filters []Filter[T]) Filter[T] {
	switch len(filters) {
	case 0:
		return none[T]
	case 1:
		return filters[0]
	}

	return func(f *FilterFS[T], op Operation) error {
		for _, filter := range filters {
			if err := filter.Matches(f.fs, op); err != nil {
				return err
			}
		}
		return nil
	}
}

func none[T fs.FS](T, Operation) error {
	return nil
}

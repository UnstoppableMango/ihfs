package ihfs

import (
	"github.com/unstoppablemango/ihfs/op"
)

type (
	FilterFunc func(*FilterFS, Operation) error
	Predicate  func(Operation) bool
)

func (p Predicate) Filter(_ *FilterFS, op Operation) error {
	if p(op) {
		return nil
	}
	return ErrPermission
}

// FilterFS is a file system that applies filter functions to operations
// before delegating them to the underlying file system.
type FilterFS struct {
	fs     FS
	filter FilterFunc
}

// Filter creates a new [FilterFS] that wraps the given file system with the provided filter functions.
func Filter(fsys FS, filters ...FilterFunc) *FilterFS {
	if fsys == nil {
		panic("filter: fsys cannot be nil")
	}

	return &FilterFS{
		fs:     fsys,
		filter: flat(filters),
	}
}

// Base implements [Decorator].
func (f *FilterFS) Base() FS {
	return f.fs
}

func (f *FilterFS) Name() string {
	return "filter"
}

// Stat implements [StatFS].
func (f *FilterFS) Stat(name string) (FileInfo, error) {
	op := op.Stat{Name: name}
	if err := f.filter(f, op); err != nil {
		return nil, err
	}
	return Stat(f.fs, name)
}

// Open implements [FS].
func (f *FilterFS) Open(name string) (File, error) {
	op := op.Open{Name: name}
	if err := f.filter(f, op); err != nil {
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
	return Filter(fsys, filters...)
}

func flat(filters []FilterFunc) FilterFunc {
	switch len(filters) {
	case 0:
		return none
	case 1:
		return filters[0]
	}

	return func(f *FilterFS, op Operation) error {
		for _, filter := range filters {
			if err := filter(f, op); err != nil {
				return err
			}
		}
		return nil
	}
}

func none(*FilterFS, Operation) error {
	return nil
}

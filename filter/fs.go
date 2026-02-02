package filter

import (
	"io/fs"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/op"
	"github.com/unstoppablemango/ihfs/try"
)

type (
	Filter    func(*FS, ihfs.Operation) error
	Predicate func(ihfs.Operation) bool
)

var None Filter = none

func (p Predicate) Filter(_ *FS, op ihfs.Operation) error {
	if p(op) {
		return nil
	}
	return ihfs.ErrPermission
}

type FS struct {
	fs     ihfs.FS
	filter Filter
}

func With(fsys ihfs.FS, filters ...Filter) *FS {
	if fsys == nil {
		panic("filter: fsys cannot be nil")
	}
	return &FS{fs: fsys, filter: flat(filters)}
}

func (f *FS) Name() string {
	return "filter"
}

// Open implements [fs.FS].
func (f *FS) Open(name string) (fs.File, error) {
	op := op.Open{Name: name}
	if err := f.filter(f, op); err != nil {
		return nil, err
	}
	return f.fs.Open(name)
}

func (f *FS) Stat(name string) (fs.FileInfo, error) {
	op := op.Stat{Name: name}
	if err := f.filter(f, op); err != nil {
		return nil, err
	}
	return try.Stat(f.fs, name)
}

func Where(fsys ihfs.FS, predicates ...Predicate) *FS {
	var filters []Filter
	for _, predicate := range predicates {
		filters = append(filters, predicate.Filter)
	}
	return With(fsys, filters...)
}

func flat(filters []Filter) Filter {
	switch len(filters) {
	case 0:
		return none
	case 1:
		return filters[0]
	}

	return func(fsys *FS, op ihfs.Operation) error {
		for _, filter := range filters {
			if err := filter(fsys, op); err != nil {
				return err
			}
		}
		return nil
	}
}

func none(_ *FS, _ ihfs.Operation) error {
	return nil
}

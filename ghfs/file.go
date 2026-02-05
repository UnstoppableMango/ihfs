package ghfs

import (
	"github.com/unstoppablemango/ihfs"
)

type Owner struct {
	name string
}

// Close implements [fs.File].
func (*Owner) Close() error {
	return nil
}

func (o *Owner) Name() string {
	return o.name
}

// Read implements [fs.File].
func (o *Owner) Read(p []byte) (n int, err error) {
	return 0, nil
}

// Stat implements [fs.File].
func (o *Owner) Stat() (ihfs.FileInfo, error) {
	return nil, nil
}

type Repository struct {
	name  string
	owner string
}

// Close implements [fs.File].
func (r *Repository) Close() error {
	return nil
}

func (r *Repository) Name() string {
	return r.name
}

func (r *Repository) Owner() string {
	return r.owner
}

// Read implements [fs.File].
func (r *Repository) Read([]byte) (int, error) {
	return 0, nil
}

// Stat implements [fs.File].
func (r *Repository) Stat() (ihfs.FileInfo, error) {
	return nil, nil
}

type Release struct {
	name       string
	owner      string
	repository string
}

// Close implements [fs.File].
func (r *Release) Close() error {
	return nil
}

func (r *Release) Name() string {
	return r.name
}

func (r *Release) Owner() string {
	return r.owner
}

func (r *Release) Repository() string {
	return r.repository
}

// Read implements [fs.File].
func (r *Release) Read([]byte) (int, error) {
	return 0, nil
}

// Stat implements [fs.File].
func (r *Release) Stat() (ihfs.FileInfo, error) {
	return nil, nil
}

type Asset struct {
	name       string
	owner      string
	repository string
	release    string
}

// Close implements [fs.File].
func (a *Asset) Close() error {
	return nil
}

func (a *Asset) Name() string {
	return a.name
}

func (a *Asset) Release() string {
	return a.release
}

func (a *Asset) Repository() string {
	return a.repository
}

func (a *Asset) Owner() string {
	return a.owner
}

// Read implements [fs.File].
func (a *Asset) Read([]byte) (int, error) {
	return 0, nil
}

// Stat implements [fs.File].
func (a *Asset) Stat() (ihfs.FileInfo, error) {
	return nil, nil
}

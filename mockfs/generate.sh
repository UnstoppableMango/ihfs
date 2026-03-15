#!/usr/bin/env bash
set -euo pipefail

fs_interfaces=(
	ChmodFS
	ChownFS
	ChtimesFS
	CopyFS
	CreateFS
	CreateTempFS
	GlobFS
	LinkerFS
	MkdirFS
	MkdirAllFS
	MkdirTempFS
	OpenFileFS
	OsFS
	ReadDirFS
	ReadDirNamesFS
	ReadFileFS
	ReadLinkFS
	RemoveFS
	RemoveAllFS
	RenameFS
	StatFS
	SubFS
	SymlinkFS
	TempFileFS
	WriteFileFS
)

file_interfaces=(
	DirNameReader
	DirReader
	File
	Operation
	ReaderAt
	ReadDirFile
	Seeker
	StringWriter
	Syncer
	Truncater
	Writer
	WriterAt
)

join()       { local IFS=,; echo "$*"; }
mock_names() { local names=(); for i in "$@"; do names+=("$i=$i"); done; join "${names[@]}"; }

mockgen \
	-destination=mock_fs.go \
	-package=mockfs \
	-mock_names="$(mock_names "${fs_interfaces[@]}")" \
	github.com/unstoppablemango/ihfs \
	"$(join "${fs_interfaces[@]}")"

mockgen \
	-destination=mock_file.go \
	-package=mockfs \
	-mock_names="$(mock_names "${file_interfaces[@]}")" \
	github.com/unstoppablemango/ihfs \
	"$(join "${file_interfaces[@]}")"

package corfs_test

import (
	"errors"
	"io"
	"io/fs"
	"testing/fstest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/corfs"
	"github.com/unstoppablemango/ihfs/errfs"
	"github.com/unstoppablemango/ihfs/memfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

type minimalFS struct{}

func (m minimalFS) Open(name string) (ihfs.File, error) {
	return nil, fs.ErrNotExist
}

func (m minimalFS) Stat(name string) (ihfs.FileInfo, error) {
	return nil, fs.ErrNotExist
}

type minimalFSWithMkdirAll struct{ minimalFS }

func (m minimalFSWithMkdirAll) MkdirAll(name string, perm ihfs.FileMode) error {
	return nil
}

var _ = Describe("Fs", func() {
	It("should return the base filesystem", func() {
		fsys := memfs.New()

		cfs := corfs.New(fsys, memfs.New())

		Expect(cfs.Base()).To(BeIdenticalTo(fsys))
	})

	It("should have a name", func() {
		cfs := corfs.New(memfs.New(), memfs.New())

		Expect(cfs.Name()).To(Equal("corfs"))
	})

	Describe("Open", func() {
		It("should cache file from base on first read", func() {
			baseContent := "base content"
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte(baseContent)), io.EOF
				},
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("test.txt")
					fi.SizeFunc = func() int64 { return int64(len(baseContent)) }
					return fi, nil
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
			)

			layerFile := &testfs.File{
				WriteFunc: func(p []byte) (int, error) {
					return len(p), nil
				},
			}

			var fileCreated bool
			layer := testfs.New(
				testfs.WithMkdirAll(func(name string, perm ihfs.FileMode) error {
					return nil
				}),
				testfs.WithCreate(func(name string) (ihfs.Writer, error) {
					fileCreated = true
					return layerFile, nil
				}),
				testfs.WithChtimes(func(name string, atime, mtime time.Time) error {
					return nil
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerFile, nil
				}),
			)

			cfs := corfs.New(base, layer)
			file, err := cfs.Open("test.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
			Expect(fileCreated).To(BeTrue(), "file should have been cached to layer")
		})

		It("should read from cache on subsequent reads", func() {
			layerFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("cached content")), io.EOF
				},
			}
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
			)

			cfs := corfs.New(testfs.New(), layer)
			file, err := cfs.Open("test.txt")
			Expect(err).ToNot(HaveOccurred())

			data, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("cached content"))
		})

		It("should open directories from base when not cached", func() {
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return &testfs.File{}, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			cfs := corfs.New(base, testfs.New())
			file, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})

		It("should return error when file doesn't exist", func() {
			base := errfs.New(fs.ErrNotExist)

			cfs := corfs.New(base, testfs.New())

			_, err := cfs.Open("nonexistent.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should handle merged directories", func() {
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return &testfs.File{}, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			cfs := corfs.New(testfs.New(), layer)
			file, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})
	})

	Describe("copyToLayer", func() {
		It("should handle directory creation in layer", func() {
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return &testfs.File{}, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
			)

			layer := testfs.New(
				testfs.WithMkdirAll(func(string, ihfs.FileMode) error {
					return errors.New("mkdirall failed")
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("subdir/test.txt")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mkdirall failed"))
		})

		It("should fail when Create is not supported", func() {
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return &testfs.File{}, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
			)

			cfs := corfs.New(base, testfs.New())
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("permission denied"))
		})

		It("should fail when write fails", func() {
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("content")), io.EOF
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
			)

			failingWriterFile := &testfs.File{
				WriteFunc: func(p []byte) (int, error) {
					return 0, errors.New("write error")
				},
			}

			var removeCalled bool
			layer := testfs.New(
				testfs.WithMkdirAll(func(name string, perm ihfs.FileMode) error {
					return nil
				}),
				testfs.WithCreate(func(name string) (ihfs.File, error) {
					return failingWriterFile, nil
				}),
				testfs.WithRemove(func(name string) error {
					removeCalled = true
					return nil
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("write error"))
			Expect(removeCalled).To(BeTrue())
		})

		It("should handle copy failure", func() {
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return &testfs.File{}, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
			)

			var removeCalled bool
			layer := testfs.New(
				testfs.WithMkdirAll(func(name string, perm ihfs.FileMode) error {
					return nil
				}),
				testfs.WithCreate(func(name string) (ihfs.File, error) {
					return &testfs.File{}, nil
				}),
				testfs.WithRemove(func(name string) error {
					removeCalled = true
					return nil
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
			Expect(removeCalled).To(BeTrue())
		})

		It("should handle cacheLocal state", func() {
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return &testfs.File{}, nil
				}),
			)

			cfs := corfs.New(memfs.New(), layer, corfs.WithCacheTime(1*time.Second))
			file, err := cfs.Open("test.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})

		It("should handle cacheStale for directory", func() {
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return &testfs.File{}, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
			)

			cfs := corfs.New(testfs.New(), layer, corfs.WithCacheTime(1*time.Hour))
			file, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})

		It("should handle base stat error when checking stale cache", func() {
			base := errfs.New(errors.New("base stat error"))

			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return &testfs.File{}, nil
				}),
			)

			cfs := corfs.New(base, layer, corfs.WithCacheTime(1*time.Hour))
			file, err := cfs.Open("test.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})

		It("should handle directory copy when layer supports MkdirAll", func() {
			baseDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("dir")
					fi.IsDirFunc = func() bool { return true }
					fi.ModeFunc = func() ihfs.FileMode { return 0755 }
					return fi, nil
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
			)

			layer := testfs.New(
				testfs.WithMkdirAll(func(string, ihfs.FileMode) error {
					return nil
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("dir")
			Expect(err).To(HaveOccurred())
		})

		It("should handle directory without MkdirAll support in copyToLayer", func() {
			baseDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("dir")
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
			)

			cfs := corfs.New(base, minimalFS{})
			_, err := cfs.Open("path")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ihfs.ErrNotImplemented))
		})

		It("should handle when layer doesn't support Create", func() {
			baseFile := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					return testfs.NewFileInfo("test.txt"), nil
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
			)

			cfs := corfs.New(base, minimalFSWithMkdirAll{})
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ihfs.ErrNotImplemented))
		})

		It("should handle layer file Close error in copyToLayer", func() {
			baseContent := []byte("test content")
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					n := copy(p, baseContent)
					return n, io.EOF
				},
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("test.txt")
					fi.SizeFunc = func() int64 { return int64(len(baseContent)) }
					return fi, nil
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
			)

			closeErr := errors.New("close error")
			layerFile := &testfs.File{
				WriteFunc: func(p []byte) (int, error) {
					return len(p), nil
				},
				CloseFunc: func() error {
					return closeErr
				},
			}
			layer := testfs.New(
				testfs.WithMkdirAll(func(name string, perm ihfs.FileMode) error {
					return nil
				}),
				testfs.WithCreate(func(name string) (ihfs.File, error) {
					return layerFile, nil
				}),
				testfs.WithChtimes(func(name string, atime, mtime time.Time) error {
					return nil
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("close error"))
		})

		It("should handle Chtimes failure gracefully", func() {
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("content")), io.EOF
				},
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("test.txt")
					fi.SizeFunc = func() int64 { return 7 }
					return fi, nil
				},
			}

			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
			)

			layerFile := &testfs.File{
				WriteFunc: func(p []byte) (int, error) {
					return len(p), nil
				},
				CloseFunc: func() error {
					return nil
				},
			}

			layer := testfs.New(
				testfs.WithMkdirAll(func(name string, perm ihfs.FileMode) error {
					return nil
				}),
				testfs.WithCreate(func(name string) (ihfs.Writer, error) {
					return layerFile, nil
				}),
				testfs.WithChtimes(func(name string, atime, mtime time.Time) error {
					return errors.New("chtimes error")
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerFile, nil
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("chtimes error"))
		})

		It("should handle cacheStatus error", func() {
			base := memfs.New()
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, errors.New("stat error")
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stat error"))
		})

		It("should merge directories when layer doesn't support MkdirAll for cacheMiss", func() {
			baseDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("dir")
					fi.IsDirFunc = func() bool { return true }
					fi.ModeFunc = func() ihfs.FileMode { return 0755 }
					return fi, nil
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			cfs := corfs.New(base, testfs.New())
			file, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})

		It("should handle cacheLocal case in Open", func() {
			base := memfs.New()

			now := time.Now()
			layerFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("local")), io.EOF
				},
			}
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.ModTimeFunc = func() time.Time { return now.Add(1 * time.Hour) }
					return fi, nil
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerFile, nil
				}),
			)

			cfs := corfs.New(base, layer, corfs.WithCacheTime(1*time.Second))
			file, err := cfs.Open("test.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})

		It("should handle layer error when base returns nil file", func() {
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return nil, fs.ErrNotExist
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			layerErr := errors.New("layer error")
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return nil, layerErr
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("dir")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(layerErr))
		})

		It("should merge directories when both base and layer errors occur", func() {
			baseDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("dir")
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return nil, errors.New("layer error")
				}),
			)

			cfs := corfs.New(base, layer)
			file, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})

		It("should handle layer error when base is nil for directory merge", func() {
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return nil, fs.ErrNotExist
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			layerDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("dir")
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
			}
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			cfs := corfs.New(base, layer)
			file, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})

		It("should handle cacheStale state for non-directory file", func() {
			base := testfs.New()

			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(name), nil
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return &testfs.File{}, nil
				}),
			)

			cfs := corfs.New(base, layer, corfs.WithCacheTime(1*time.Hour))
			file, err := cfs.Open("test.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})

		It("should handle cacheStale for directory", func() {
			now := time.Now()
			oldTime := now.Add(-2 * time.Hour)

			baseDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("dir")
					fi.IsDirFunc = func() bool { return true }
					fi.ModTimeFunc = func() time.Time { return now }
					return fi, nil
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return true }
					fi.ModTimeFunc = func() time.Time { return now }
					return fi, nil
				}),
			)

			layerDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("dir")
					fi.IsDirFunc = func() bool { return true }
					fi.ModTimeFunc = func() time.Time { return oldTime }
					return fi, nil
				},
			}
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return true }
					fi.ModTimeFunc = func() time.Time { return oldTime }
					return fi, nil
				}),
			)

			cfs := corfs.New(base, layer, corfs.WithCacheTime(1*time.Hour))
			file, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})
	})

	Describe("WriteFile", func() {
		It("should write through to base and layer", func() {
			var baseWritten, layerWritten string
			base := testfs.New(
				testfs.WithWriteFile(func(name string, _ []byte, _ ihfs.FileMode) error {
					baseWritten = name
					return nil
				}),
			)
			layer := testfs.New(
				testfs.WithWriteFile(func(name string, _ []byte, _ ihfs.FileMode) error {
					layerWritten = name
					return nil
				}),
			)

			cfs := corfs.New(base, layer)
			err := cfs.WriteFile("test.txt", []byte("data"), 0644)
			Expect(err).NotTo(HaveOccurred())
			Expect(baseWritten).To(Equal("test.txt"))
			Expect(layerWritten).To(Equal("test.txt"))
		})

		It("should return error when base does not support WriteFile", func() {
			cfs := corfs.New(&testfs.BoringFs{}, testfs.New())
			err := cfs.WriteFile("test.txt", []byte("data"), 0644)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when layer does not support WriteFile", func() {
			cfs := corfs.New(testfs.New(), &testfs.BoringFs{})
			err := cfs.WriteFile("test.txt", []byte("data"), 0644)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("fstest", func() {
		It("should pass fstest.TestFS", func() {
			base := memfs.New()
			layer := memfs.New()

			Expect(base.Mkdir("dir", 0755)).To(Succeed())

			f, err := base.Create("file.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = f.(io.Writer).Write([]byte("content"))
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			f2, err := base.Create("dir/nested.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = f2.(io.Writer).Write([]byte("nested"))
			Expect(err).NotTo(HaveOccurred())
			Expect(f2.Close()).To(Succeed())

			cfs := corfs.New(base, layer)

			err = fstest.TestFS(cfs, "file.txt", "dir", "dir/nested.txt")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

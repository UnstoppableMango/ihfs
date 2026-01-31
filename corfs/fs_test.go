package corfs_test

import (
	"errors"
	"io"
	"io/fs"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/corfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

// nonCreateFS is a filesystem that doesn't implement CreateFS
type nonCreateFS struct {
	testfs.Fs
}

// Override Stat to not implement CreateFS
func (f nonCreateFS) Create(name string) (ihfs.File, error) {
	panic("should not be called - this FS doesn't implement CreateFS")
}

// nonWriterFile is a file that doesn't implement Writer
type nonWriterFile struct {
	*testfs.File
}

// Override Write to not satisfy Writer interface signature properly
func (f *nonWriterFile) Write(p []byte) (int, error) {
	return 0, errors.New("write not supported")
}

var _ = Describe("Fs", func() {
	Describe("Open", func() {
		It("should cache file from base on first read", func() {
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("base content")), io.EOF
				},
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("test.txt")
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return time.Now() }
					fi.ModeFunc = func() ihfs.FileMode { return 0644 }
					fi.SizeFunc = func() int64 { return 12 }
					return fi, nil
				},
				CloseFunc: func() error { return nil },
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return time.Now() }
					return fi, nil
				}),
			)

			layerFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("base content")), io.EOF
				},
				WriteFunc: func(p []byte) (int, error) {
					return len(p), nil
				},
				CloseFunc: func() error { return nil },
			}

			var fileCreated bool
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					if fileCreated {
						fi := testfs.NewFileInfo(name)
						fi.IsDirFunc = func() bool { return false }
						fi.ModTimeFunc = func() time.Time { return time.Now() }
						return fi, nil
					}
					return nil, fs.ErrNotExist
				}),
				testfs.WithCreate(func(name string) (ihfs.File, error) {
					fileCreated = true
					return layerFile, nil
				}),
				testfs.WithChtimes(func(name string, atime, mtime time.Time) error {
					return nil
				}),
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					if fileCreated {
						return layerFile, nil
					}
					return nil, fs.ErrNotExist
				}),
			)

			cfs := corfs.New(base, layer)
			file, err := cfs.Open("test.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
			Expect(fileCreated).To(BeTrue(), "file should have been cached to layer")
		})

		It("should read from cache on subsequent reads", func() {
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("base content")), io.EOF
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
			)

			layerFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("cached content")), io.EOF
				},
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("not applicable")
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return time.Now() }
					return fi, nil
				},
			}
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return time.Now() }
					return fi, nil
				}),
			)

			cfs := corfs.New(base, layer)
			file, err := cfs.Open("test.txt")
			Expect(err).ToNot(HaveOccurred())

			buf := make([]byte, 100)
			n, _ := file.Read(buf)
			Expect(string(buf[:n])).To(Equal("cached content"))
		})

		It("should open directories from base when not cached", func() {
			baseDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("not applicable")
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
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
				}),
			)

			cfs := corfs.New(base, layer)
			file, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})

		It("should return error when file doesn't exist", func() {
			base := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
				}),
			)
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("nonexistent.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should respect cache time", func() {
			now := time.Now()
			oldTime := now.Add(-2 * time.Hour)

			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("new content")), io.EOF
				},
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("not applicable")
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return now }
					fi.ModeFunc = func() ihfs.FileMode { return 0644 }
					return fi, nil
				},
				CloseFunc: func() error { return nil },
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return now }
					return fi, nil
				}),
			)

			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return oldTime }
					return fi, nil
				}),
			)

			cfs := corfs.New(base, layer, corfs.WithCacheTime(1*time.Hour))
			_, err := cfs.Open("test.txt")
			// This will fail to copy because layer doesn't support Create
			Expect(err).To(HaveOccurred())
		})

		It("should handle merged directories", func() {
			baseDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("not applicable")
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
					fi.ModTimeFunc = func() time.Time { return time.Now() }
					return fi, nil
				}),
			)

			layerDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("not applicable")
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
					fi.ModTimeFunc = func() time.Time { return time.Now() }
					return fi, nil
				}),
			)

			cfs := corfs.New(base, layer)
			file, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})
	})

	Describe("cacheStatus", func() {
		It("should return cacheMiss when file not in layer", func() {
			base := testfs.New()
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
				}),
			)

			cfs := corfs.New(base, layer)
			// Access internal method through Open behavior
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should return cacheHit with zero cache time", func() {
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("base")), io.EOF
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
			)

			layerFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("cached")), io.EOF
				},
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("not applicable")
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return time.Now() }
					return fi, nil
				},
			}
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return time.Now() }
					return fi, nil
				}),
			)

			cfs := corfs.New(base, layer) // Zero cache time
			file, err := cfs.Open("test.txt")
			Expect(err).ToNot(HaveOccurred())

			buf := make([]byte, 100)
			n, _ := file.Read(buf)
			Expect(string(buf[:n])).To(Equal("cached"))
		})
	})

	Describe("copyToLayer", func() {
		It("should handle copy errors", func() {
			baseFile := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					return nil, errors.New("stat error")
				},
				CloseFunc: func() error { return nil },
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					return fi, nil
				}),
			)
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should handle directory creation in layer", func() {
			// Test MkdirAll error when creating parent directories for a file
			baseFile := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("test.txt")
					fi.IsDirFunc = func() bool { return false }
					fi.ModeFunc = func() ihfs.FileMode { return 0644 }
					fi.SizeFunc = func() int64 { return 10 }
					return fi, nil
				},
				CloseFunc: func() error { return nil },
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					return fi, nil
				}),
			)

			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
				}),
				testfs.WithMkdirAll(func(path string, perm ihfs.FileMode) error {
					return errors.New("mkdirall failed")
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("subdir/test.txt")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mkdirall failed"))
		})

		It("should fail when Create is not supported", func() {
			baseFile := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("test.txt")
					fi.IsDirFunc = func() bool { return false }
					fi.ModeFunc = func() ihfs.FileMode { return 0644 }
					fi.SizeFunc = func() int64 { return 10 }
					return fi, nil
				},
				CloseFunc: func() error { return nil },
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					return fi, nil
				}),
			)

			// Use default testfs which returns permission error from Create
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create layer file"))
		})

		It("should fail when write fails", func() {
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("content")), io.EOF
				},
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("test.txt")
					fi.IsDirFunc = func() bool { return false }
					fi.ModeFunc = func() ihfs.FileMode { return 0644 }
					fi.SizeFunc = func() int64 { return 7 }
					return fi, nil
				},
				CloseFunc: func() error { return nil },
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					return fi, nil
				}),
			)

			// Create a file with a failing Write
			failingWriterFile := &testfs.File{
				WriteFunc: func(p []byte) (int, error) {
					return 0, errors.New("write error")
				},
				CloseFunc: func() error { return nil },
			}

			var removeCalled bool
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
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
			Expect(err.Error()).To(ContainSubstring("failed to copy file contents"))
			Expect(removeCalled).To(BeTrue())
		})

		It("should handle copy failure", func() {
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return 0, errors.New("read error")
				},
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("test.txt")
					fi.IsDirFunc = func() bool { return false }
					fi.ModeFunc = func() ihfs.FileMode { return 0644 }
					fi.SizeFunc = func() int64 { return 10 }
					return fi, nil
				},
				CloseFunc: func() error { return nil },
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					return fi, nil
				}),
			)

			layerFile := &testfs.File{
				WriteFunc: func(p []byte) (int, error) {
					return len(p), nil
				},
				CloseFunc: func() error { return nil },
			}

			var removeCalled bool
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
				}),
				testfs.WithCreate(func(name string) (ihfs.File, error) {
					return layerFile, nil
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

		It("should handle error opening base file", func() {
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return nil, errors.New("open error")
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					return fi, nil
				}),
			)
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should handle cacheLocal state", func() {
			base := testfs.New()

			layerFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("local")), io.EOF
				},
				CloseFunc: func() error { return nil },
			}
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return time.Now().Add(1 * time.Hour) }
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
				CloseFunc: func() error { return nil },
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
				CloseFunc: func() error { return nil },
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

		It("should handle base error when opening merged directory", func() {
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return nil, errors.New("base error")
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return nil, fs.ErrNotExist
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("dir")
			Expect(err).To(HaveOccurred())
		})

		It("should handle error creating parent directories", func() {
			baseFile := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo("test.txt")
					fi.IsDirFunc = func() bool { return false }
					fi.ModeFunc = func() ihfs.FileMode { return 0644 }
					fi.SizeFunc = func() int64 { return 10 }
					return fi, nil
				},
				CloseFunc: func() error { return nil },
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo(name)
					fi.IsDirFunc = func() bool { return false }
					return fi, nil
				}),
			)

			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
				}),
				testfs.WithMkdirAll(func(path string, perm ihfs.FileMode) error {
					return errors.New("mkdirall error")
				}),
			)

			cfs := corfs.New(base, layer)
			_, err := cfs.Open("dir/test.txt")
			Expect(err).To(HaveOccurred())
		})
	})
})

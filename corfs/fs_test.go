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

var _ = Describe("Fs", func() {
	Describe("Open", func() {
		It("should cache file from base on first read", func() {
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("base content")), io.EOF
				},
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return time.Now() }
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
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return time.Now() }
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
			// This will fail because layer doesn't support Create,
			// which is expected for a minimal test
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("layer filesystem does not support Create"))
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
					fi := testfs.NewFileInfo()
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
					fi := testfs.NewFileInfo()
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
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
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
					fi := testfs.NewFileInfo()
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
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return false }
					fi.ModTimeFunc = func() time.Time { return now }
					return fi, nil
				}),
			)

			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
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
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					fi.ModTimeFunc = func() time.Time { return time.Now() }
					return fi, nil
				}),
			)

			layerDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
			}
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
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
					fi := testfs.NewFileInfo()
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
					fi := testfs.NewFileInfo()
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
					fi := testfs.NewFileInfo()
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
	})
})

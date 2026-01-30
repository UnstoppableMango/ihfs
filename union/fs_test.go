package union_test

import (
	"errors"
	"io"
	"io/fs"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/cowfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("Fs", func() {
	Describe("Open", func() {
		It("should open file from base", func() {
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("base")), io.EOF
				},
			}
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(), nil
				}),
			)
			layer := testfs.New()

			cfs := cowfs.New(base, layer)
			file, err := cfs.Open("test.txt")
			Expect(err).ToNot(HaveOccurred())

			buf := make([]byte, 100)
			n, _ := file.Read(buf)
			Expect(string(buf[:n])).To(Equal("base"))
		})

		It("should open file from layer", func() {
			layerFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("layer")), io.EOF
				},
			}
			base := testfs.New()
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return testfs.NewFileInfo(), nil
				}),
			)

			cfs := cowfs.New(base, layer)
			file, err := cfs.Open("test.txt")
			Expect(err).ToNot(HaveOccurred())

			buf := make([]byte, 100)
			n, _ := file.Read(buf)
			Expect(string(buf[:n])).To(Equal("layer"))
		})

		It("should merge directories from both layers", func() {
			baseDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
			}
			layerDir := &testfs.File{
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
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			cfs := cowfs.New(base, layer)
			file, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())
			Expect(file).ToNot(BeNil())
		})

		It("should return error when base stat fails", func() {
			base := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, errors.New("stat error")
				}),
			)
			layer := testfs.New()

			cfs := cowfs.New(base, layer)
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when layer stat fails", func() {
			base := testfs.New()
			layer := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, errors.New("stat error")
				}),
			)

			cfs := cowfs.New(base, layer)
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should return joined error when base open fails", func() {
			layerDir := &testfs.File{
				CloseFunc: func() error { return nil },
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
			}

			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return nil, errors.New("open error")
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			cfs := cowfs.New(base, layer)
			_, err := cfs.Open("dir")
			Expect(err).To(HaveOccurred())
			var pathErr *ihfs.PathError
			Expect(errors.As(err, &pathErr)).To(BeTrue())
			Expect(pathErr.Op).To(Equal("open"))
			Expect(pathErr.Path).To(Equal("dir"))
		})

		It("should return joined error when layer open fails", func() {
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
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return nil, errors.New("open error")
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			cfs := cowfs.New(base, layer)
			_, err := cfs.Open("dir")
			Expect(err).To(HaveOccurred())
			var pathErr *ihfs.PathError
			Expect(errors.As(err, &pathErr)).To(BeTrue())
			Expect(pathErr.Op).To(Equal("open"))
			Expect(pathErr.Path).To(Equal("dir"))
		})

		It("should return error when file doesn't exist", func() {
			cfs := cowfs.New(testfs.New(), testfs.New())
			_, err := cfs.Open("nonexistent.txt")
			Expect(err).To(HaveOccurred())
		})

		It("should open layer directory when base is not directory", func() {
			layerDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
			}

			base := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return false }
					return fi, nil
				}),
			)
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			cfs := cowfs.New(base, layer)
			_, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("isInBase", func() {
		It("should handle ErrNotExist", func() {
			base := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, fs.ErrNotExist
				}),
			)

			cfs := cowfs.New(base, testfs.New())
			_, err := cfs.Open("test.txt")
			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("should handle ENOENT", func() {
			base := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, syscall.ENOENT
				}),
			)

			cfs := cowfs.New(base, testfs.New())
			_, err := cfs.Open("test.txt")
			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("should handle ENOTDIR", func() {
			base := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, syscall.ENOTDIR
				}),
			)

			cfs := cowfs.New(base, testfs.New())
			_, err := cfs.Open("test.txt")
			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		It("should return other errors", func() {
			base := testfs.New(
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					return nil, errors.New("other error")
				}),
			)

			cfs := cowfs.New(base, testfs.New())
			_, err := cfs.Open("test.txt")
			Expect(err).To(HaveOccurred())
		})
	})
})

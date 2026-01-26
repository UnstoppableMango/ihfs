package cowfs_test

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
		Context("when file exists only in base", func() {
			It("should open file from base", func() {
				content := []byte("base content")
				baseFile := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						n := copy(p, content)
						return n, io.EOF
					},
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.NameFunc = func() string { return "test.txt" }
						return fi, nil
					},
				}

				base := testfs.New(
					testfs.WithOpen(func(name string) (ihfs.File, error) {
						if name == "test.txt" {
							return baseFile, nil
						}
						return nil, fs.ErrNotExist
					}),
					testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
						if name == "test.txt" {
							fi := testfs.NewFileInfo()
							fi.NameFunc = func() string { return "test.txt" }
							return fi, nil
						}
						return nil, fs.ErrNotExist
					}),
				)
				layer := testfs.New()

				cfs := cowfs.New(base, layer)
				file, err := cfs.Open("test.txt")
				Expect(err).ToNot(HaveOccurred())
				Expect(file).ToNot(BeNil())
				defer file.Close()

				buf := make([]byte, 100)
				n, _ := file.Read(buf)
				Expect(string(buf[:n])).To(Equal("base content"))
			})
		})

		Context("when file exists only in layer", func() {
			It("should open file from layer", func() {
				content := []byte("layer content")
				layerFile := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						n := copy(p, content)
						return n, io.EOF
					},
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.NameFunc = func() string { return "test.txt" }
						fi.IsDirFunc = func() bool { return false }
						return fi, nil
					},
				}

				base := testfs.New()
				layer := testfs.New(
					testfs.WithOpen(func(name string) (ihfs.File, error) {
						if name == "test.txt" {
							return layerFile, nil
						}
						return nil, fs.ErrNotExist
					}),
					testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
						if name == "test.txt" {
							fi := testfs.NewFileInfo()
							fi.NameFunc = func() string { return "test.txt" }
							fi.IsDirFunc = func() bool { return false }
							return fi, nil
						}
						return nil, fs.ErrNotExist
					}),
				)

				cfs := cowfs.New(base, layer)
				file, err := cfs.Open("test.txt")
				Expect(err).ToNot(HaveOccurred())
				Expect(file).ToNot(BeNil())
				defer file.Close()

				buf := make([]byte, 100)
				n, _ := file.Read(buf)
				Expect(string(buf[:n])).To(Equal("layer content"))
			})
		})

		Context("when directory exists in both layers", func() {
			It("should merge both directories", func() {
				baseDir := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.NameFunc = func() string { return "dir" }
						fi.IsDirFunc = func() bool { return true }
						return fi, nil
					},
				}
				layerDir := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.NameFunc = func() string { return "dir" }
						fi.IsDirFunc = func() bool { return true }
						return fi, nil
					},
				}

				base := testfs.New(
					testfs.WithOpen(func(name string) (ihfs.File, error) {
						if name == "dir" {
							return baseDir, nil
						}
						return nil, fs.ErrNotExist
					}),
					testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
						if name == "dir" {
							fi := testfs.NewFileInfo()
							fi.IsDirFunc = func() bool { return true }
							return fi, nil
						}
						return nil, fs.ErrNotExist
					}),
				)
				layer := testfs.New(
					testfs.WithOpen(func(name string) (ihfs.File, error) {
						if name == "dir" {
							return layerDir, nil
						}
						return nil, fs.ErrNotExist
					}),
					testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
						if name == "dir" {
							fi := testfs.NewFileInfo()
							fi.IsDirFunc = func() bool { return true }
							return fi, nil
						}
						return nil, fs.ErrNotExist
					}),
				)

				cfs := cowfs.New(base, layer)
				file, err := cfs.Open("dir")
				Expect(err).ToNot(HaveOccurred())
				Expect(file).ToNot(BeNil())
				defer file.Close()
			})
		})

		Context("when base returns error on isInBase check", func() {
			It("should return the error", func() {
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
		})

		Context("when layer directory check returns error", func() {
			It("should return the error", func() {
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
		})

		Context("when opening directory and base returns error", func() {
			It("should return joined error", func() {
				layerDir := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
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
						if name == "dir" {
							return layerDir, nil
						}
						return nil, fs.ErrNotExist
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
		})

		Context("when opening directory and layer returns error", func() {
			It("should return joined error", func() {
				baseDir := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.IsDirFunc = func() bool { return true }
						return fi, nil
					},
				}

				base := testfs.New(
					testfs.WithOpen(func(name string) (ihfs.File, error) {
						if name == "dir" {
							return baseDir, nil
						}
						return nil, fs.ErrNotExist
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
		})

		Context("when file doesn't exist", func() {
			It("should return error", func() {
				base := testfs.New()
				layer := testfs.New()

				cfs := cowfs.New(base, layer)
				_, err := cfs.Open("nonexistent.txt")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when base is not a directory but layer is", func() {
			It("should return layer directory", func() {
				layerDir := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.IsDirFunc = func() bool { return true }
						return fi, nil
					},
				}

				base := testfs.New(
					testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.IsDirFunc = func() bool { return false } // Not a directory
						return fi, nil
					}),
				)
				layer := testfs.New(
					testfs.WithOpen(func(name string) (ihfs.File, error) {
						if name == "dir" {
							return layerDir, nil
						}
						return nil, fs.ErrNotExist
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
				defer file.Close()
			})
		})
	})

	Describe("isInBase", func() {
		Context("when base stat returns ErrNotExist", func() {
			It("should return false with no error", func() {
				base := testfs.New(
					testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
						return nil, fs.ErrNotExist
					}),
				)
				layer := testfs.New()

				cfs := cowfs.New(base, layer)
				_, err := cfs.Open("test.txt")
				Expect(err).To(HaveOccurred()) // Should get error from layer too
			})
		})

		Context("when base stat returns ENOENT", func() {
			It("should return false with no error", func() {
				base := testfs.New(
					testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
						return nil, syscall.ENOENT
					}),
				)
				layer := testfs.New()

				cfs := cowfs.New(base, layer)
				_, err := cfs.Open("test.txt")
				Expect(err).To(HaveOccurred()) // Should get error from layer too
			})
		})

		Context("when base stat returns ENOTDIR", func() {
			It("should return false with no error", func() {
				base := testfs.New(
					testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
						return nil, syscall.ENOTDIR
					}),
				)
				layer := testfs.New()

				cfs := cowfs.New(base, layer)
				_, err := cfs.Open("test.txt")
				Expect(err).To(HaveOccurred()) // Should get error from layer too
			})
		})

		Context("when base stat returns other error", func() {
			It("should return the error", func() {
				base := testfs.New(
					testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
						return nil, errors.New("other error")
					}),
				)
				layer := testfs.New()

				cfs := cowfs.New(base, layer)
				_, err := cfs.Open("test.txt")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

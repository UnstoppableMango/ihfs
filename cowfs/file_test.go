package cowfs_test

import (
	"errors"
	"io"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/cowfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("File", func() {
	Describe("NewFile", func() {
		It("should create a new File", func() {
			baseFile := &testfs.File{
				CloseFunc: func() error { return nil },
				ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					return fi, nil
				},
			}
			layerFile := &testfs.File{
				CloseFunc: func() error { return nil },
				ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					return fi, nil
				},
			}

			file := cowfs.NewFile("test", baseFile, layerFile)
			Expect(file).ToNot(BeNil())
		})
	})

	Describe("Close", func() {
		Context("when both base and layer are present", func() {
			It("should close both files", func() {
				closed := make(map[string]bool)
				baseDir := &testfs.File{
					CloseFunc: func() error {
						closed["base"] = true
						return nil
					},
					ReadFunc: func(p []byte) (int, error) { return 0, io.EOF },
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.IsDirFunc = func() bool { return true }
						return fi, nil
					},
				}
				layerDir := &testfs.File{
					CloseFunc: func() error {
						closed["layer"] = true
						return nil
					},
					ReadFunc: func(p []byte) (int, error) { return 0, io.EOF },
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

				err = file.Close()
				Expect(err).ToNot(HaveOccurred())
				Expect(closed["base"]).To(BeTrue())
				Expect(closed["layer"]).To(BeTrue())
			})
		})

		Context("when only layer is present", func() {
			It("should close layer file", func() {
				layerFile := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						return copy(p, []byte("layer")), io.EOF
					},
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
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
						fi := testfs.NewFileInfo()
						fi.IsDirFunc = func() bool { return false }
						return fi, nil
					}),
				)

				cfs := cowfs.New(base, layer)
				file, err := cfs.Open("test.txt")
				Expect(err).ToNot(HaveOccurred())

				err = file.Close()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when only base is present", func() {
			It("should close base file", func() {
				baseFile := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						return copy(p, []byte("base")), io.EOF
					},
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
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
						fi := testfs.NewFileInfo()
						return fi, nil
					}),
				)
				layer := testfs.New()

				cfs := cowfs.New(base, layer)
				file, err := cfs.Open("test.txt")
				Expect(err).ToNot(HaveOccurred())

				err = file.Close()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when neither base nor layer exists (using NewFile)", func() {
			It("should return BADFD error", func() {
				file := cowfs.NewFile("test", nil, nil)
				err := file.Close()
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(cowfs.BADFD))
			})
		})
	})

	Describe("Read", func() {
		Context("when reading from both layers", func() {
			It("should read from layer and sync base position", func() {
				baseDir := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.IsDirFunc = func() bool { return true }
						return fi, nil
					},
					SeekFunc: func(offset int64, whence int) (int64, error) {
						return offset, nil
					},
				}
				layerDir := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return copy(p, []byte("layer")), nil },
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
				defer file.Close()

				buf := make([]byte, 5)
				n, err := file.Read(buf)
				Expect(err).To(SatisfyAny(
					BeNil(),
					Equal(io.EOF),
				))
				Expect(n).To(BeNumerically(">=", 0))
			})
		})

		Context("when layer is nil but base exists", func() {
			It("should read from base", func() {
				baseFile := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						return copy(p, []byte("base content")), io.EOF
					},
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
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
						fi := testfs.NewFileInfo()
						return fi, nil
					}),
				)
				layer := testfs.New()

				cfs := cowfs.New(base, layer)
				file, err := cfs.Open("test.txt")
				Expect(err).ToNot(HaveOccurred())
				defer file.Close()

				buf := make([]byte, 100)
				n, err := file.Read(buf)
				Expect(err).To(SatisfyAny(
					BeNil(),
					Equal(io.EOF),
				))
				Expect(string(buf[:n])).To(Equal("base content"))
			})
		})

		Context("when layer returns EOF", func() {
			It("should sync base position", func() {
				baseDir := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.IsDirFunc = func() bool { return true }
						return fi, nil
					},
					SeekFunc: func(offset int64, whence int) (int64, error) {
						return offset, nil
					},
				}
				layerDir := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return 1, io.EOF },
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
				defer file.Close()

				buf := make([]byte, 100)
				n, err := file.Read(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(1))
			})
		})

		Context("when base seek fails", func() {
			It("should return seek error", func() {
				seekErr := errors.New("seek failed")
				baseDir := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.IsDirFunc = func() bool { return true }
						return fi, nil
					},
					SeekFunc: func(offset int64, whence int) (int64, error) {
						return 0, seekErr
					},
				}
				layerDir := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return copy(p, []byte("data")), nil },
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
				defer file.Close()

				buf := make([]byte, 100)
				n, err := file.Read(buf)
				Expect(n).To(Equal(4))
				Expect(err).To(Equal(seekErr))
			})
		})

		Context("when both layer and base are nil (using NewFile)", func() {
			It("should return BADFD error", func() {
				file := cowfs.NewFile("test", nil, nil)
				buf := make([]byte, 100)
				n, err := file.Read(buf)
				Expect(n).To(Equal(0))
				Expect(err).To(Equal(cowfs.BADFD))
			})
		})

		Context("when only base exists (using NewFile)", func() {
			It("should read from base", func() {
				baseFile := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						return copy(p, []byte("from base")), io.EOF
					},
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						return fi, nil
					},
				}

				file := cowfs.NewFile("test", baseFile, nil)
				buf := make([]byte, 100)
				n, err := file.Read(buf)
				Expect(err).To(SatisfyAny(BeNil(), Equal(io.EOF)))
				Expect(string(buf[:n])).To(Equal("from base"))
			})
		})

		Context("when layer read errors (not EOF)", func() {
			It("should return the error", func() {
				readErr := errors.New("read error")
				layerFile := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						return 0, readErr
					},
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						return fi, nil
					},
				}

				file := cowfs.NewFile("test", nil, layerFile)
				buf := make([]byte, 100)
				_, err := file.Read(buf)
				Expect(err).To(Equal(readErr))
			})
		})
	})

	Describe("Stat", func() {
		Context("when cowfs File with layer exists", func() {
			It("should return layer file info", func() {
				baseDir := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.NameFunc = func() string { return "dir" }
						fi.IsDirFunc = func() bool { return true }
						fi.SizeFunc = func() int64 { return 100 }
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
						fi.SizeFunc = func() int64 { return 200 }
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
				defer file.Close()

				info, err := file.Stat()
				Expect(err).ToNot(HaveOccurred())
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal("dir"))
				Expect(info.Size()).To(Equal(int64(200))) // Should be from layer
			})
		})

		Context("when layer file exists (non-cowfs File)", func() {
			It("should return layer file info", func() {
				layerFile := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						return copy(p, []byte("layer")), io.EOF
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
						fi := testfs.NewFileInfo()
						fi.NameFunc = func() string { return "test.txt" }
						fi.IsDirFunc = func() bool { return false }
						return fi, nil
					}),
				)

				cfs := cowfs.New(base, layer)
				file, err := cfs.Open("test.txt")
				Expect(err).ToNot(HaveOccurred())
				defer file.Close()

				info, err := file.Stat()
				Expect(err).ToNot(HaveOccurred())
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal("test.txt"))
			})
		})

		Context("when only base file exists", func() {
			It("should return base file info", func() {
				baseFile := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						return copy(p, []byte("base")), io.EOF
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
						fi := testfs.NewFileInfo()
						fi.NameFunc = func() string { return "test.txt" }
						return fi, nil
					}),
				)
				layer := testfs.New()

				cfs := cowfs.New(base, layer)
				file, err := cfs.Open("test.txt")
				Expect(err).ToNot(HaveOccurred())
				defer file.Close()

				info, err := file.Stat()
				Expect(err).ToNot(HaveOccurred())
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal("test.txt"))
			})
		})

		Context("when only base exists (using NewFile)", func() {
			It("should return base file info", func() {
				baseFile := &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.NameFunc = func() string { return "test.txt" }
						return fi, nil
					},
				}

				file := cowfs.NewFile("test", baseFile, nil)
				info, err := file.Stat()
				Expect(err).ToNot(HaveOccurred())
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal("test.txt"))
			})
		})

		Context("when neither layer nor base exists (using NewFile)", func() {
			It("should return BADFD error", func() {
				file := cowfs.NewFile("test", nil, nil)
				info, err := file.Stat()
				Expect(info).To(BeNil())
				Expect(err).To(Equal(cowfs.BADFD))
			})
		})
	})
})

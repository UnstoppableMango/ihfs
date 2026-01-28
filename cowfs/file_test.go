package cowfs_test

import (
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/cowfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("File", func() {
	Describe("Close", func() {
		Context("when both base and layer are present", func() {
			It("should close both files", func() {
				var base, layer bool
				file := cowfs.NewFile(
					&testfs.File{
						CloseFunc: func() error {
							base = true
							return nil
						},
					},
					&testfs.File{
						CloseFunc: func() error {
							layer = true
							return nil
						},
					},
				)

				err := file.Close()

				Expect(err).ToNot(HaveOccurred())
				Expect(base).To(BeTrue())
				Expect(layer).To(BeTrue())
			})
		})

		Context("when only layer is present", func() {
			It("should close layer file", func() {
				file := cowfs.NewFile(nil, &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						return 0, io.EOF
					},
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.IsDirFunc = func() bool { return false }
						return fi, nil
					},
				})

				Expect(file.Close()).NotTo(HaveOccurred())
			})
		})

		Context("when neither base nor layer exists", func() {
			It("should return BADFD error", func() {
				file := cowfs.NewFile(nil, nil)

				err := file.Close()

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(cowfs.BADFD))
			})
		})
	})

	Describe("Read", func() {
		Context("when reading from both layers", func() {
			// TODO: This test does not actually verify that the base position is synced.
			It("should read from layer and sync base position", func() {
				file := cowfs.NewFile(
					&testfs.File{
						SeekFunc: func(offset int64, whence int) (int64, error) {
							return offset, nil
						},
					},
					&testfs.File{
						SeekFunc: func(offset int64, whence int) (int64, error) {
							return offset, nil
						},
					},
				)

				buf := make([]byte, 5)
				_, err := file.Read(buf)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when layer is nil but base exists", func() {
			It("should read from base", func() {
				file := cowfs.NewFile(&testfs.File{
					ReadFunc: func(p []byte) (int, error) {
						return copy(p, []byte("base content")), io.EOF
					},
				}, nil)

				buf := make([]byte, 100)
				n, err := file.Read(buf)
				Expect(err).To(MatchError(io.EOF))
				Expect(string(buf[:n])).To(Equal("base content"))
			})
		})

		Context("when layer returns EOF", func() {
			It("should sync base position", func() {
				file := cowfs.NewFile(
					&testfs.File{
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
					},
					&testfs.File{
						CloseFunc: func() error { return nil },
						ReadFunc:  func(p []byte) (int, error) { return 1, io.EOF },
						StatFunc: func() (ihfs.FileInfo, error) {
							fi := testfs.NewFileInfo()
							fi.IsDirFunc = func() bool { return true }
							return fi, nil
						},
					},
				)

				buf := make([]byte, 100)
				n, err := file.Read(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(1))
			})
		})

		Context("when base seek fails", func() {
			It("should return seek error", func() {
				seekErr := errors.New("seek failed")
				file := cowfs.NewFile(
					&testfs.File{
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
					},
					&testfs.File{
						CloseFunc: func() error { return nil },
						ReadFunc:  func(p []byte) (int, error) { return copy(p, []byte("data")), nil },
						StatFunc: func() (ihfs.FileInfo, error) {
							fi := testfs.NewFileInfo()
							fi.IsDirFunc = func() bool { return true }
							return fi, nil
						},
					},
				)

				buf := make([]byte, 100)
				n, err := file.Read(buf)
				Expect(n).To(Equal(4))
				Expect(err).To(Equal(seekErr))
			})
		})

		Context("when both layer and base are nil", func() {
			It("should return BADFD error", func() {
				file := cowfs.NewFile(nil, nil)
				buf := make([]byte, 100)
				n, err := file.Read(buf)
				Expect(n).To(Equal(0))
				Expect(err).To(Equal(cowfs.BADFD))
			})
		})

		Context("when only base exists", func() {
			It("should read from base", func() {
				file := cowfs.NewFile(&testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						return copy(p, []byte("from base")), io.EOF
					},
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						return fi, nil
					},
				}, nil)

				buf := make([]byte, 100)
				n, err := file.Read(buf)
				Expect(err).To(SatisfyAny(BeNil(), Equal(io.EOF)))
				Expect(string(buf[:n])).To(Equal("from base"))
			})
		})

		Context("when layer read errors (not EOF)", func() {
			It("should return the error", func() {
				readErr := errors.New("read error")
				file := cowfs.NewFile(nil, &testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						return 0, readErr
					},
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						return fi, nil
					},
				})

				buf := make([]byte, 100)
				_, err := file.Read(buf)
				Expect(err).To(Equal(readErr))
			})
		})
	})

	Describe("Stat", func() {
		Context("when cowfs File with layer exists", func() {
			It("should return layer file info", func() {
				file := cowfs.NewFile(
					&testfs.File{
						CloseFunc: func() error { return nil },
						ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
						StatFunc: func() (ihfs.FileInfo, error) {
							fi := testfs.NewFileInfo()
							fi.NameFunc = func() string { return "dir" }
							fi.IsDirFunc = func() bool { return true }
							fi.SizeFunc = func() int64 { return 100 }
							return fi, nil
						},
					},
					&testfs.File{
						CloseFunc: func() error { return nil },
						ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
						StatFunc: func() (ihfs.FileInfo, error) {
							fi := testfs.NewFileInfo()
							fi.NameFunc = func() string { return "dir" }
							fi.IsDirFunc = func() bool { return true }
							fi.SizeFunc = func() int64 { return 200 }
							return fi, nil
						},
					},
				)

				info, err := file.Stat()
				Expect(err).ToNot(HaveOccurred())
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal("dir"))
				Expect(info.Size()).To(Equal(int64(200))) // Should be from layer
			})
		})

		Context("when layer file exists (non-cowfs File)", func() {
			It("should return layer file info", func() {
				file := cowfs.NewFile(nil, &testfs.File{
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
				})

				info, err := file.Stat()
				Expect(err).ToNot(HaveOccurred())
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal("test.txt"))
			})
		})

		Context("when only base file exists", func() {
			It("should return base file info", func() {
				file := cowfs.NewFile(&testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc: func(p []byte) (int, error) {
						return copy(p, []byte("base")), io.EOF
					},
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.NameFunc = func() string { return "test.txt" }
						return fi, nil
					},
				}, nil)

				info, err := file.Stat()
				Expect(err).ToNot(HaveOccurred())
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal("test.txt"))
			})
		})

		Context("when only base exists", func() {
			It("should return base file info", func() {
				file := cowfs.NewFile(&testfs.File{
					CloseFunc: func() error { return nil },
					ReadFunc:  func(p []byte) (int, error) { return 0, io.EOF },
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.NameFunc = func() string { return "test.txt" }
						return fi, nil
					},
				}, nil)

				info, err := file.Stat()
				Expect(err).ToNot(HaveOccurred())
				Expect(info).ToNot(BeNil())
				Expect(info.Name()).To(Equal("test.txt"))
			})
		})

		Context("when neither layer nor base exists", func() {
			It("should return BADFD error", func() {
				file := cowfs.NewFile(nil, nil)
				info, err := file.Stat()
				Expect(info).To(BeNil())
				Expect(err).To(Equal(cowfs.BADFD))
			})
		})
	})
})

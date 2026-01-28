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

			Expect(err).NotTo(HaveOccurred())
			Expect(base).To(BeTrue())
			Expect(layer).To(BeTrue())
		})

		It("should close layer only", func() {
			var called bool
			file := cowfs.NewFile(nil, &testfs.File{
				CloseFunc: func() error {
					called = true
					return nil
				},
			})

			Expect(file.Close()).NotTo(HaveOccurred())
			Expect(called).To(BeTrue())
		})

		It("should return base errors", func() {
			baseErr := errors.New("base close error")
			file := cowfs.NewFile(
				&testfs.File{
					CloseFunc: func() error {
						return baseErr
					},
				},
				&testfs.File{
					CloseFunc: func() error {
						return nil
					},
				},
			)

			err := file.Close()
			Expect(err).To(MatchError("base close error"))
		})

		It("should return layer errors", func() {
			layerErr := errors.New("layer close error")
			file := cowfs.NewFile(
				&testfs.File{
					CloseFunc: func() error {
						return nil
					},
				},
				&testfs.File{
					CloseFunc: func() error {
						return layerErr
					},
				},
			)

			err := file.Close()
			Expect(err).To(MatchError("layer close error"))
		})

		It("should return BADFD when neither exists", func() {
			file := cowfs.NewFile(nil, nil)
			Expect(file.Close()).To(Equal(cowfs.BADFD))
		})
	})

	Describe("Read", func() {
		It("should read from layer", func() {
			file := cowfs.NewFile(nil, &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("layer")), io.EOF
				},
			})

			buf := make([]byte, 100)
			n, err := file.Read(buf)
			Expect(err).To(MatchError(io.EOF))
			Expect(string(buf[:n])).To(Equal("layer"))
		})

		It("should read from base", func() {
			file := cowfs.NewFile(&testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("base")), io.EOF
				},
			}, nil)

			buf := make([]byte, 100)
			n, err := file.Read(buf)
			Expect(err).To(MatchError(io.EOF))
			Expect(string(buf[:n])).To(Equal("base"))
		})

		It("should sync base position on read", func() {
			var seekOffset int64
			var seekWhence int
			file := cowfs.NewFile(
				&testfs.File{
					SeekFunc: func(offset int64, whence int) (int64, error) {
						seekOffset = offset
						seekWhence = whence
						return offset, nil
					},
				},
				&testfs.File{
					ReadFunc: func(p []byte) (int, error) {
						return 1, io.EOF
					},
				},
			)

			buf := make([]byte, 100)
			_, _ = file.Read(buf)
			Expect(seekOffset).To(Equal(int64(1)))
			Expect(seekWhence).To(Equal(io.SeekCurrent))
		})

		It("should return seek error", func() {
			seekErr := errors.New("seek failed")
			file := cowfs.NewFile(
				&testfs.File{
					SeekFunc: func(offset int64, whence int) (int64, error) {
						return 0, seekErr
					},
				},
				&testfs.File{
					ReadFunc: func(p []byte) (int, error) {
						return copy(p, []byte("data")), nil
					},
				},
			)

			buf := make([]byte, 100)
			n, err := file.Read(buf)
			Expect(n).To(Equal(4))
			Expect(err).To(Equal(seekErr))
		})

		It("should return BADFD when neither exists", func() {
			file := cowfs.NewFile(nil, nil)
			buf := make([]byte, 100)
			n, err := file.Read(buf)
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(cowfs.BADFD))
		})

		It("should return read error", func() {
			readErr := errors.New("read error")
			file := cowfs.NewFile(nil, &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return 0, readErr
				},
			})

			buf := make([]byte, 100)
			_, err := file.Read(buf)
			Expect(err).To(Equal(readErr))
		})
	})

	Describe("Stat", func() {
		It("should return layer info", func() {
			file := cowfs.NewFile(
				&testfs.File{
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.SizeFunc = func() int64 { return 100 }
						return fi, nil
					},
				},
				&testfs.File{
					StatFunc: func() (ihfs.FileInfo, error) {
						fi := testfs.NewFileInfo()
						fi.SizeFunc = func() int64 { return 200 }
						return fi, nil
					},
				},
			)

			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Size()).To(Equal(int64(200)))
		})

		It("should return base info", func() {
			file := cowfs.NewFile(&testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.NameFunc = func() string { return "base.txt" }
					return fi, nil
				},
			}, nil)

			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Name()).To(Equal("base.txt"))
		})

		It("should return BADFD when neither exists", func() {
			file := cowfs.NewFile(nil, nil)
			info, err := file.Stat()
			Expect(info).To(BeNil())
			Expect(err).To(Equal(cowfs.BADFD))
		})
	})
})

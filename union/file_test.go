package union_test

import (
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/fsutil/try"
	"github.com/unstoppablemango/ihfs/testfs"
	"github.com/unstoppablemango/ihfs/union"
)

var _ = Describe("File", func() {
	Describe("Close", func() {
		It("should close both files", func() {
			var base, layer bool
			file := union.NewFile(
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
			file := union.NewFile(nil, &testfs.File{
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
			file := union.NewFile(
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
			file := union.NewFile(
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
			file := union.NewFile(nil, nil)
			Expect(file.Close()).To(Equal(union.BADFD))
		})
	})

	Describe("Read", func() {
		It("should read from layer", func() {
			file := union.NewFile(nil, &testfs.File{
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
			file := union.NewFile(&testfs.File{
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
			file := union.NewFile(
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
			file := union.NewFile(
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
			file := union.NewFile(nil, nil)
			buf := make([]byte, 100)
			n, err := file.Read(buf)
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(union.BADFD))
		})

		It("should return read error", func() {
			readErr := errors.New("read error")
			file := union.NewFile(nil, &testfs.File{
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
			file := union.NewFile(
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
			file := union.NewFile(&testfs.File{
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
			file := union.NewFile(nil, nil)
			info, err := file.Stat()
			Expect(info).To(BeNil())
			Expect(err).To(Equal(union.BADFD))
		})
	})

	Describe("ReadDir", func() {
		It("should merge entries from both layers", func() {
			baseEntry := testfs.NewDirEntry("base.txt", false)
			layerEntry := testfs.NewDirEntry("layer.txt", false)
			sharedEntry := testfs.NewDirEntry("shared.txt", false)

			baseFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{baseEntry, sharedEntry}, nil
				},
			}
			layerFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{layerEntry, sharedEntry}, nil
				},
			}

			file := union.NewFile(baseFile, layerFile)
			entries, err := file.ReadDir(-1)

			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(3))

			names := make(map[string]bool)
			for _, e := range entries {
				names[e.Name()] = true
			}
			Expect(names).To(HaveKey("base.txt"))
			Expect(names).To(HaveKey("layer.txt"))
			Expect(names).To(HaveKey("shared.txt"))
		})

		It("should prioritize layer over base for duplicates", func() {
			sharedEntry := testfs.NewDirEntry("shared.txt", false)

			baseFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{sharedEntry}, nil
				},
			}
			layerFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{sharedEntry}, nil
				},
			}

			file := union.NewFile(baseFile, layerFile)
			entries, err := file.ReadDir(-1)

			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("shared.txt"))
		})

		It("should support pagination", func() {
			entry1 := testfs.NewDirEntry("file1.txt", false)
			entry2 := testfs.NewDirEntry("file2.txt", false)

			layerFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry1, entry2}, nil
				},
			}
			file := union.NewFile(nil, layerFile)

			entries, err := file.ReadDir(1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))

			entries, err = file.ReadDir(1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))

			entries, err = file.ReadDir(1)
			Expect(err).To(Equal(io.EOF))
			Expect(entries).To(BeNil())
		})

		It("should return EOF when pagination exhausted", func() {
			entry := testfs.NewDirEntry("file.txt", false)
			layerFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry}, nil
				},
			}
			file := union.NewFile(nil, layerFile)

			_, _ = file.ReadDir(1)
			entries, err := file.ReadDir(1)

			Expect(err).To(Equal(io.EOF))
			Expect(entries).To(BeNil())
		})

		It("should return error when layer ReadDir fails", func() {
			layerErr := errors.New("layer readdir failed")
			layerFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return nil, layerErr
				},
			}
			file := union.NewFile(nil, layerFile)

			entries, err := file.ReadDir(-1)

			Expect(err).To(Equal(layerErr))
			Expect(entries).To(BeNil())
		})

		It("should return error when base ReadDir fails", func() {
			baseErr := errors.New("base readdir failed")
			baseFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return nil, baseErr
				},
			}
			file := union.NewFile(baseFile, nil)

			entries, err := file.ReadDir(-1)

			Expect(err).To(Equal(baseErr))
			Expect(entries).To(BeNil())
		})

		It("should handle layer not implementing ReadDirFile", func() {
			baseEntry := testfs.NewDirEntry("base.txt", false)
			baseFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{baseEntry}, nil
				},
			}
			layerFile := &testfs.BoringFile{}
			file := union.NewFile(baseFile, layerFile)

			entries, err := file.ReadDir(-1)

			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("base.txt"))
		})

		It("should handle base not implementing ReadDirFile", func() {
			layerEntry := testfs.NewDirEntry("layer.txt", false)
			layerFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{layerEntry}, nil
				},
			}
			baseFile := &testfs.BoringFile{}
			file := union.NewFile(baseFile, layerFile)

			entries, err := file.ReadDir(-1)

			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("layer.txt"))
		})

		It("should handle neither implementing ReadDirFile", func() {
			file := union.NewFile(&testfs.BoringFile{}, &testfs.BoringFile{})

			entries, err := file.ReadDir(-1)

			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(BeEmpty())
		})

		It("should return remaining entries when n exceeds available", func() {
			entry1 := testfs.NewDirEntry("file1.txt", false)
			entry2 := testfs.NewDirEntry("file2.txt", false)
			entry3 := testfs.NewDirEntry("file3.txt", false)

			layerFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry1, entry2, entry3}, nil
				},
			}
			file := union.NewFile(nil, layerFile)

			// Read 2 entries first
			entries, err := file.ReadDir(2)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(2))

			// Request 10 entries but only 1 remains
			entries, err = file.ReadDir(10)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
		})
	})

	Describe("Write", func() {
		It("should write to layer and base when both exist", func() {
			var baseData, layerData []byte
			file := union.NewFile(
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						baseData = append(baseData, p...)
						return len(p), nil
					},
				},
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						layerData = append(layerData, p...)
						return len(p), nil
					},
				},
			)

			n, err := file.Write([]byte("data"))
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(4))
			Expect(string(layerData)).To(Equal("data"))
			Expect(string(baseData)).To(Equal("data"))
		})

		It("should write to base only", func() {
			var baseData []byte
			file := union.NewFile(&testfs.File{
				WriteFunc: func(p []byte) (int, error) {
					baseData = append(baseData, p...)
					return len(p), nil
				},
			}, nil)

			n, err := file.Write([]byte("base data"))
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(9))
			Expect(string(baseData)).To(Equal("base data"))
		})

		It("should write to layer when layer write fails with base", func() {
			writeErr := errors.New("layer error")
			baseErr := errors.New("base error")
			file := union.NewFile(
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						return 0, baseErr
					},
				},
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						return 0, writeErr
					},
				},
			)

			n, err := file.Write([]byte("both"))
			Expect(err).To(Equal(baseErr))
			Expect(n).To(Equal(0))
		})

		It("should return layer byte count", func() {
			file := union.NewFile(
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						return 10, nil
					},
				},
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						return 5, nil
					},
				},
			)

			n, err := file.Write([]byte("test"))
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(5))
		})

		It("should return layer write error when base not present", func() {
			writeErr := errors.New("layer write failed")
			file := union.NewFile(nil, &testfs.File{
				WriteFunc: func(p []byte) (int, error) {
					return 0, writeErr
				},
			})

			n, err := file.Write([]byte("data"))
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(writeErr))
		})

		It("should return base write error when layer succeeds", func() {
			baseErr := errors.New("base write failed")
			file := union.NewFile(
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						return 0, baseErr
					},
				},
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						return len(p), nil
					},
				},
			)

			n, err := file.Write([]byte("data"))
			Expect(n).To(Equal(4))
			Expect(err).To(Equal(baseErr))
		})

		It("should write to base when layer succeeds and base exists", func() {
			var baseCalled bool
			file := union.NewFile(
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						baseCalled = true
						return len(p), nil
					},
				},
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						return len(p), nil
					},
				},
			)

			n, err := file.Write([]byte("data"))
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(4))
			Expect(baseCalled).To(BeTrue())
		})

		It("should return BADFD when neither exists", func() {
			file := union.NewFile(nil, nil)
			n, err := file.Write([]byte("data"))
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(union.BADFD))
		})

		It("should handle empty writes", func() {
			var layerData, baseData []byte
			file := union.NewFile(
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						baseData = append(baseData, p...)
						return len(p), nil
					},
				},
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						layerData = append(layerData, p...)
						return len(p), nil
					},
				},
			)

			n, err := file.Write([]byte{})
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(0))
			Expect(layerData).To(BeEmpty())
			Expect(baseData).To(BeEmpty())
		})

		It("should handle layer not supporting write with base", func() {
			baseErr := errors.New("base error")
			file := union.NewFile(
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						return 0, baseErr
					},
				},
				&testfs.BoringFile{},
			)

			n, err := file.Write([]byte("data"))
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(baseErr))
		})

		It("should handle base not supporting write when layer succeeds", func() {
			var layerData []byte
			file := union.NewFile(
				&testfs.BoringFile{},
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						layerData = append(layerData, p...)
						return len(p), nil
					},
				},
			)

			n, err := file.Write([]byte("test"))
			Expect(n).To(Equal(4))
			Expect(string(layerData)).To(Equal("test"))
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})

		It("should handle base not supporting write when only base exists", func() {
			file := union.NewFile(&testfs.BoringFile{}, nil)

			n, err := file.Write([]byte("data"))
			Expect(n).To(Equal(0))
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})

		It("should handle large writes", func() {
			var layerSize, baseSize int
			largeData := make([]byte, 10*1024) // 10 KB
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			file := union.NewFile(
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						baseSize += len(p)
						return len(p), nil
					},
				},
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						layerSize += len(p)
						return len(p), nil
					},
				},
			)

			n, err := file.Write(largeData)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))
			Expect(layerSize).To(Equal(len(largeData)))
			Expect(baseSize).To(Equal(len(largeData)))
		})

		It("should handle partial layer writes", func() {
			file := union.NewFile(
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						return len(p), nil
					},
				},
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						return len(p) / 2, nil
					},
				},
			)

			n, err := file.Write([]byte("testdata"))
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(4))
		})

		It("should handle partial base writes", func() {
			file := union.NewFile(
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						return len(p) / 2, nil
					},
				},
				&testfs.File{
					WriteFunc: func(p []byte) (int, error) {
						return len(p), nil
					},
				},
			)

			n, err := file.Write([]byte("testdata"))
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(8))
		})
	})
})

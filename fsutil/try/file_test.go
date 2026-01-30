package try_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/fsutil/try"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("File", func() {
	Describe("Seek", func() {
		It("should return an error when file does not support seeking", func() {
			f := &testfs.BoringFile{}

			_, err := try.Seek(f, 0, 0)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})

		It("should call Seek on the underlying file when supported", func() {
			var actualOffset int64
			var actualWhence int

			f := &testfs.File{
				SeekFunc: func(offset int64, whence int) (int64, error) {
					actualOffset = offset
					actualWhence = whence
					return 69, errors.New("test error")
				},
			}

			n, err := try.Seek(f, 420, 67)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("test error"))
			Expect(n).To(Equal(int64(69)))
			Expect(actualOffset).To(Equal(int64(420)))
			Expect(actualWhence).To(Equal(67))
		})
	})

	Describe("Write", func() {
		It("should return an error when file does not support writing", func() {
			f := &testfs.BoringFile{}

			_, err := try.Write(f, nil)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})

		It("should call Write on the underlying file when supported", func() {
			var actualData []byte

			f := &testfs.File{
				WriteFunc: func(p []byte) (int, error) {
					actualData = p
					return 69, errors.New("test error")
				},
			}

			n, err := try.Write(f, []byte("hello"))

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("test error"))
			Expect(n).To(Equal(69))
			Expect(actualData).To(Equal([]byte("hello")))
		})
	})

	Describe("ReadAt", func() {
		It("should return an error when file does not support ReadAt", func() {
			f := &testfs.BoringFile{}

			_, err := try.ReadAt(f, nil, 0)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})

		It("should call ReadAt on the underlying file when supported", func() {
			var actualData []byte
			var actualOff int64

			f := &mockReaderAt{
				readAt: func(p []byte, off int64) (int, error) {
					actualData = p
					actualOff = off
					return 42, errors.New("test error")
				},
			}

			buf := make([]byte, 10)
			n, err := try.ReadAt(f, buf, 100)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("test error"))
			Expect(n).To(Equal(42))
			Expect(actualData).To(Equal(buf))
			Expect(actualOff).To(Equal(int64(100)))
		})
	})

	Describe("WriteAt", func() {
		It("should return an error when file does not support WriteAt", func() {
			f := &testfs.BoringFile{}

			_, err := try.WriteAt(f, nil, 0)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})

		It("should call WriteAt on the underlying file when supported", func() {
			var actualData []byte
			var actualOff int64

			f := &mockWriterAt{
				writeAt: func(p []byte, off int64) (int, error) {
					actualData = p
					actualOff = off
					return 42, errors.New("test error")
				},
			}

			n, err := try.WriteAt(f, []byte("hello"), 100)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("test error"))
			Expect(n).To(Equal(42))
			Expect(actualData).To(Equal([]byte("hello")))
			Expect(actualOff).To(Equal(int64(100)))
		})
	})

	Describe("WriteString", func() {
		It("should return an error when file does not support WriteString", func() {
			f := &testfs.BoringFile{}

			_, err := try.WriteString(f, "hello")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})

		It("should call WriteString on the underlying file when supported", func() {
			var actualString string

			f := &mockStringWriter{
				writeString: func(s string) (int, error) {
					actualString = s
					return 42, errors.New("test error")
				},
			}

			n, err := try.WriteString(f, "hello")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("test error"))
			Expect(n).To(Equal(42))
			Expect(actualString).To(Equal("hello"))
		})
	})

	Describe("Sync", func() {
		It("should return an error when file does not support Sync", func() {
			f := &testfs.BoringFile{}

			err := try.Sync(f)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})

		It("should call Sync on the underlying file when supported", func() {
			var called bool

			f := &mockSyncer{
				sync: func() error {
					called = true
					return errors.New("test error")
				},
			}

			err := try.Sync(f)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("test error"))
			Expect(called).To(BeTrue())
		})
	})

	Describe("Truncate", func() {
		It("should return an error when file does not support Truncate", func() {
			f := &testfs.BoringFile{}

			err := try.Truncate(f, 100)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})

		It("should call Truncate on the underlying file when supported", func() {
			var actualSize int64

			f := &mockTruncater{
				truncate: func(size int64) error {
					actualSize = size
					return errors.New("test error")
				},
			}

			err := try.Truncate(f, 100)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("test error"))
			Expect(actualSize).To(Equal(int64(100)))
		})
	})

	Describe("ReadDirFile", func() {
		It("should return an error when file does not support ReadDir", func() {
			f := &testfs.BoringFile{}

			_, err := try.ReadDirFile(f, 0)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})

		It("should call ReadDir on the underlying file when supported", func() {
			var actualN int

			entry := testfs.NewDirEntry("file.txt", false)
			f := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					actualN = n
					return []ihfs.DirEntry{entry}, errors.New("test error")
				},
			}

			entries, err := try.ReadDirFile(f, 10)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("test error"))
			Expect(entries).To(HaveLen(1))
			Expect(entries[0]).To(Equal(entry))
			Expect(actualN).To(Equal(10))
		})
	})

	Describe("ReadDirNamesFile", func() {
		It("should return an error when file does not support ReadDirNames", func() {
			f := &testfs.BoringFile{}

			_, err := try.ReadDirNamesFile(f, 0)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(try.ErrNotSupported))
		})

		It("should call ReadDirNames on the underlying file when supported", func() {
			var actualN int

			f := &mockDirNameReader{
				readDirNames: func(n int) ([]string, error) {
					actualN = n
					return []string{"file1.txt", "file2.txt"}, errors.New("test error")
				},
			}

			names, err := try.ReadDirNamesFile(f, 10)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("test error"))
			Expect(names).To(Equal([]string{"file1.txt", "file2.txt"}))
			Expect(actualN).To(Equal(10))
		})
	})
})

// Mock implementations for file interfaces

type mockReaderAt struct {
	testfs.BoringFile
	readAt func(p []byte, off int64) (int, error)
}

func (m *mockReaderAt) ReadAt(p []byte, off int64) (int, error) {
	return m.readAt(p, off)
}

type mockWriterAt struct {
	testfs.BoringFile
	writeAt func(p []byte, off int64) (int, error)
}

func (m *mockWriterAt) WriteAt(p []byte, off int64) (int, error) {
	return m.writeAt(p, off)
}

type mockStringWriter struct {
	testfs.BoringFile
	writeString func(s string) (int, error)
}

func (m *mockStringWriter) WriteString(s string) (int, error) {
	return m.writeString(s)
}

type mockSyncer struct {
	testfs.BoringFile
	sync func() error
}

func (m *mockSyncer) Sync() error {
	return m.sync()
}

type mockTruncater struct {
	testfs.BoringFile
	truncate func(size int64) error
}

func (m *mockTruncater) Truncate(size int64) error {
	return m.truncate(size)
}

type mockDirNameReader struct {
	testfs.BoringFile
	readDirNames func(n int) ([]string, error)
}

func (m *mockDirNameReader) ReadDirNames(n int) ([]string, error) {
	return m.readDirNames(n)
}

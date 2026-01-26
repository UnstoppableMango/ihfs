package try_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
})

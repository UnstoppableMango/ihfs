package ihfs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/osfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("Fs", func() {
	It("should read directory entry names", func() {
		fsys := osfs.New()

		names, err := ihfs.ReadDirNames(fsys, "./testdata/2-files")

		Expect(err).NotTo(HaveOccurred())
		Expect(names).To(ConsistOf("one.txt", "two.txt"))
	})

	It("should return error when directory does not exist", func() {
		fsys := osfs.New()

		names, err := ihfs.ReadDirNames(fsys, "./nonexistent")

		Expect(err).To(HaveOccurred())
		Expect(names).To(BeNil())
	})

	It("should use fs.ReadDir fallback when FS does not implement ReadDir", func() {
		fsys := testfs.New()

		names, err := ihfs.ReadDirNames(fsys, "./testdata/2-files")

		Expect(err).To(HaveOccurred())
		Expect(names).To(BeNil())
	})
})

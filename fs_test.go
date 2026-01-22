package ihfs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/osfs"
)

var _ = Describe("Fs", func() {
	It("should read directory entry names", func() {
		fsys := osfs.New()

		names, err := ihfs.ReadDirNames(fsys, "./testdata/2-files")

		Expect(err).NotTo(HaveOccurred())
		Expect(names).To(ConsistOf("one.txt", "two.txt"))
	})
})

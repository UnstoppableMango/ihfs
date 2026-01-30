package union_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/testfs"
	"github.com/unstoppablemango/ihfs/union"
)

var _ = Describe("MergeStrategy", func() {
	Describe("Ordering", func() {
		It("should maintain consistent ordering across multiple calls", func() {
			entry1 := testfs.NewDirEntry("alpha.txt", false)
			entry2 := testfs.NewDirEntry("beta.txt", false)
			entry3 := testfs.NewDirEntry("gamma.txt", false)
			entry4 := testfs.NewDirEntry("delta.txt", false)

			baseFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry1, entry3}, nil
				},
			}
			layerFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{entry2, entry4}, nil
				},
			}

			// Call ReadDir multiple times and verify consistent ordering
			var firstOrder []string
			for i := range 5 {
				file := union.NewFile(baseFile, layerFile)
				entries, err := file.ReadDir(-1)

				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(HaveLen(4))

				currentOrder := make([]string, len(entries))
				for j, e := range entries {
					currentOrder[j] = e.Name()
				}

				if i == 0 {
					firstOrder = currentOrder
				} else {
					Expect(currentOrder).To(Equal(firstOrder),
						"ReadDir should return entries in the same order across calls")
				}
			}
		})
	})
})

package union_test

import (
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/cowfs"
	"github.com/unstoppablemango/ihfs/testfs"
	"github.com/unstoppablemango/ihfs/union"
)

var _ = Describe("MergeStrategy", func() {
	Describe("WithMergeStrategy", func() {
		It("should use custom merge strategy", func() {
			customMergeCalled := false
			customMerge := func(layer, base []ihfs.DirEntry) ([]ihfs.DirEntry, error) {
				customMergeCalled = true
				return append(layer, base...), nil
			}

			baseEntry := testfs.NewDirEntry("base.txt", false)
			layerEntry := testfs.NewDirEntry("layer.txt", false)

			baseDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{baseEntry}, nil
				},
			}
			layerDir := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{layerEntry}, nil
				},
			}

			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerDir, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			cfs := cowfs.New(base, layer, cowfs.WithMergeStrategy(customMerge))
			file, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())

			dirFile, ok := file.(fs.ReadDirFile)
			Expect(ok).To(BeTrue())

			entries, err := dirFile.ReadDir(-1)
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(2))
			Expect(customMergeCalled).To(BeTrue())
		})
	})

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

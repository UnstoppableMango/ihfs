package cowfs_test

import (
	"errors"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/cowfs"
	"github.com/unstoppablemango/ihfs/testfs"
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

	Describe("WithDefaultMergeStrategy", func() {
		It("should use default merge strategy", func() {
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

			cfs := cowfs.New(base, layer, cowfs.WithDefaultMergeStrategy())
			file, err := cfs.Open("dir")
			Expect(err).ToNot(HaveOccurred())

			dirFile, ok := file.(fs.ReadDirFile)
			Expect(ok).To(BeTrue())

			entries, err := dirFile.ReadDir(-1)
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(2))
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
				file := cowfs.NewFile(baseFile, layerFile)
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

	Describe("Error Handling", func() {
		It("should return error when merge strategy fails", func() {
			mergeErr := errors.New("merge failed")
			failingMerge := func(layer, base []ihfs.DirEntry) ([]ihfs.DirEntry, error) {
				return nil, mergeErr
			}

			baseEntry := testfs.NewDirEntry("base.txt", false)
			layerEntry := testfs.NewDirEntry("layer.txt", false)

			baseFile := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{baseEntry}, nil
				},
			}
			layerFile := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				},
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{layerEntry}, nil
				},
			}

			// Test through Fs.Open which uses newFile internally with merge strategy
			base := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return baseFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)
			layer := testfs.New(
				testfs.WithOpen(func(name string) (ihfs.File, error) {
					return layerFile, nil
				}),
				testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.IsDirFunc = func() bool { return true }
					return fi, nil
				}),
			)

			// Create a cowfs with a failing merge strategy
			cfs := cowfs.New(base, layer, cowfs.WithMergeStrategy(failingMerge))
			dir, err := cfs.Open("dir")
			Expect(err).NotTo(HaveOccurred())

			dirFile, ok := dir.(fs.ReadDirFile)
			Expect(ok).To(BeTrue())

			entries, err := dirFile.ReadDir(-1)
			Expect(err).To(Equal(mergeErr))
			Expect(entries).To(BeNil())
		})
	})
})

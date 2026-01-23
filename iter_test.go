package ihfs_test

import (
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go/slices"
	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/osfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("Iter", func() {
	It("should return open errors", func() {
		fsys := testfs.New(testfs.WithOpen(func(name string) (ihfs.File, error) {
			return nil, fs.ErrNotExist
		}))

		seq := ihfs.IterPaths(fsys, "/nonexistent")

		paths, errors := slices.Collect2(seq)

		Expect(paths).To(ConsistOf("/nonexistent"))
		Expect(errors).To(ConsistOf(fs.ErrNotExist))
	})

	It("should cancel iteration when yield returns false", func() {
		fsys := osfs.New()

		seq := ihfs.IterPaths(fsys, "./testdata/2-files")

		collectedPaths := []string{}
		collectedErrors := []error{}
		for path, err := range seq {
			collectedPaths = append(collectedPaths, path)
			collectedErrors = append(collectedErrors, err)
			if len(collectedPaths) == 2 {
				break
			}
		}

		Expect(collectedErrors).To(ConsistOf(nil, nil))
		Expect(collectedPaths).To(HaveExactElements(
			"./testdata/2-files",
			"testdata/2-files/one.txt",
		))
	})

	It("should iterate over file paths", func() {
		seq := ihfs.IterPaths(osfs.New(), "./testdata/2-files")

		paths, errors := slices.Collect2(seq)

		Expect(errors).To(ConsistOf(nil, nil, nil))
		Expect(paths).To(HaveExactElements(
			"./testdata/2-files",
			"testdata/2-files/one.txt",
			"testdata/2-files/two.txt",
		))
	})

	It("should iterate over dir entries", func() {
		seq := ihfs.IterDirEntries(osfs.New(), "./testdata/2-files")

		errors := []error{}
		names := []string{}
		for d, err := range seq {
			errors = append(errors, err)
			names = append(names, d.Name())
		}

		Expect(errors).To(ConsistOf(nil, nil, nil))
		Expect(names).To(ConsistOf(
			"2-files",
			"one.txt",
			"two.txt",
		))
	})

	It("should cancel dir entries iteration when yield returns false", func() {
		fsys := osfs.New()

		seq := ihfs.IterDirEntries(fsys, "./testdata/2-files")

		collectedNames := []string{}
		collectedErrors := []error{}
		for d, err := range seq {
			collectedNames = append(collectedNames, d.Name())
			collectedErrors = append(collectedErrors, err)
			if len(collectedNames) == 2 {
				break
			}
		}

		Expect(collectedErrors).To(ConsistOf(nil, nil))
		Expect(collectedNames).To(HaveExactElements(
			"2-files",
			"one.txt",
		))
	})

	It("should iterate over paths and dir entries", func() {
		seq := ihfs.Iter(osfs.New(), "./testdata/2-files")

		paths, entries, errors := slices.Collect3(seq)

		Expect(errors).To(ConsistOf(nil, nil, nil))
		Expect(paths).To(HaveExactElements(
			"./testdata/2-files",
			"testdata/2-files/one.txt",
			"testdata/2-files/two.txt",
		))
		Expect(entries).To(HaveLen(3))
	})

	It("should cancel iteration when yield returns false in Iter", func() {
		fsys := osfs.New()

		seq := ihfs.Iter(fsys, "./testdata/2-files")

		paths := []string{}
		names := []string{}
		errors := []error{}

		seq(func(path string, d ihfs.DirEntry, err error) bool {
			paths = append(paths, path)
			names = append(names, d.Name())
			errors = append(errors, err)

			return len(paths) < 2
		})

		Expect(errors).To(ConsistOf(nil, nil))
		Expect(paths).To(HaveExactElements(
			"./testdata/2-files",
			"testdata/2-files/one.txt",
		))
		Expect(names).To(HaveExactElements(
			"2-files",
			"one.txt",
		))
	})
})

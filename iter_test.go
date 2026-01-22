package ihfs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/unmango/go/slices"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/osfs"
)

var _ = Describe("Iter", func() {
	It("should iterate over file paths", func() {
		seq := ihfs.IterPaths(osfs.New(), "./testdata/2-files")

		errors := []error{}
		paths := []string{}
		for p, err := range seq {
			errors = append(errors, err)
			paths = append(paths, p)
		}

		Expect(errors).To(ConsistOf(nil, nil, nil))
		Expect(paths).To(ConsistOf(
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

	It("should iterate over paths and dir entries", func() {
		seq := ihfs.Iter(osfs.New(), "./testdata/2-files")

		paths, entries, errors := slices.Collect3(seq)

		Expect(errors).To(ConsistOf(nil, nil, nil))
		Expect(paths).To(ConsistOf(
			"./testdata/2-files",
			"testdata/2-files/one.txt",
			"testdata/2-files/two.txt",
		))
		Expect(entries).To(ConsistOf(
			"2-files",
			"one.txt",
			"two.txt",
		))
	})
})

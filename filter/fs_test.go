package filter_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/filter"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("Fs", func() {
	It("should have a name", func() {
		fsys := filter.With(nil)

		Expect(fsys.Name()).To(Equal("filter"))
	})

	It("should pass through opens without filters", func() {
		file := &testfs.BoringFile{}
		fsys := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
			return file, nil
		}))

		filtered := filter.With(fsys)

		f, err := filtered.Open("somefile.txt")
		Expect(err).To(BeNil())
		Expect(f).To(BeIdenticalTo(file))
	})

	Context("Filter", func() {
		It("should fail to open a filtered file", func() {
			file := &testfs.BoringFile{}
			fsys := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
				return file, nil
			}))

			filtered := filter.With(fsys, func(f *filter.FS, o ihfs.Operation) error {
				if o.Subject() == "forbidden.txt" {
					return ihfs.ErrPermission
				}
				return nil
			})

			_, err := filtered.Open("forbidden.txt")
			Expect(err).To(MatchError(ihfs.ErrPermission))
			f, err := filtered.Open("allowed.txt")
			Expect(err).To(BeNil())
			Expect(f).To(BeIdenticalTo(file))
		})

		It("should apply multiple filters", func() {
			file := &testfs.BoringFile{}
			fsys := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
				return file, nil
			}))

			filtered := filter.With(fsys,
				func(f *filter.FS, o ihfs.Operation) error {
					if o.Subject() == "forbidden1.txt" {
						return ihfs.ErrPermission
					}
					return nil
				},
				func(f *filter.FS, o ihfs.Operation) error {
					if o.Subject() == "forbidden2.txt" {
						return ihfs.ErrPermission
					}
					return nil
				},
			)

			_, err := filtered.Open("forbidden1.txt")
			Expect(err).To(MatchError(ihfs.ErrPermission))
			_, err = filtered.Open("forbidden2.txt")
			Expect(err).To(MatchError(ihfs.ErrPermission))
			f, err := filtered.Open("allowed.txt")
			Expect(err).To(BeNil())
			Expect(f).To(BeIdenticalTo(file))
		})
	})

	Context("Predicate", func() {
		It("should fail to open a filtered file", func() {
			file := &testfs.BoringFile{}
			fsys := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
				return file, nil
			}))

			filtered := filter.Where(fsys, func(o ihfs.Operation) bool {
				return o.Subject() != "forbidden.txt"
			})

			_, err := filtered.Open("forbidden.txt")
			Expect(err).To(MatchError(ihfs.ErrPermission))
			f, err := filtered.Open("allowed.txt")
			Expect(err).To(BeNil())
			Expect(f).To(BeIdenticalTo(file))
		})

		It("should apply multiple predicates", func() {
			file := &testfs.BoringFile{}
			fsys := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
				return file, nil
			}))

			filtered := filter.Where(fsys,
				func(o ihfs.Operation) bool {
					return o.Subject() != "forbidden1.txt"
				},
				func(o ihfs.Operation) bool {
					return o.Subject() != "forbidden2.txt"
				},
			)

			_, err := filtered.Open("forbidden1.txt")
			Expect(err).To(MatchError(ihfs.ErrPermission))
			_, err = filtered.Open("forbidden2.txt")
			Expect(err).To(MatchError(ihfs.ErrPermission))
			f, err := filtered.Open("allowed.txt")
			Expect(err).To(BeNil())
			Expect(f).To(BeIdenticalTo(file))
		})
	})
})

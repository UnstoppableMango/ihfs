package ihfs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("Fs", func() {
	It("should panic on nil fsys", func() {
		Expect(func() {
			ihfs.Filter(nil)
		}).To(Panic())
	})

	It("should have a name", func() {
		fsys := ihfs.Filter(testfs.BoringFs{})

		Expect(fsys.Name()).To(Equal("filter"))
	})

	It("should return the base fsys", func() {
		base := &testfs.BoringFs{}
		fsys := ihfs.Filter(base)

		Expect(fsys.Base()).To(BeIdenticalTo(base))
	})

	It("should call Stat on the underlying filesystem", func() {
		called := false
		info := testfs.NewFileInfo("test.txt")
		fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
			called = true
			Expect(name).To(Equal("test.txt"))
			return info, nil
		}))

		filtered := ihfs.Filter(fsys)
		result, err := filtered.Stat("test.txt")

		Expect(err).To(BeNil())
		Expect(result).To(BeIdenticalTo(info))
		Expect(called).To(BeTrue())
	})

	It("should apply filters to Stat", func() {
		info := testfs.NewFileInfo("allowed.txt")
		fsys := testfs.New(testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
			return info, nil
		}))

		filtered := ihfs.Filter(fsys, func(f *ihfs.FilterFS, o ihfs.Operation) error {
			if o.Subject() == "forbidden.txt" {
				return ihfs.ErrPermission
			}
			return nil
		})

		_, err := filtered.Stat("forbidden.txt")
		Expect(err).To(MatchError(ihfs.ErrPermission))
		result, err := filtered.Stat("allowed.txt")
		Expect(err).To(BeNil())
		Expect(result).To(BeIdenticalTo(info))
	})

	It("should pass through opens without filters", func() {
		file := &testfs.BoringFile{}
		fsys := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
			return file, nil
		}))

		filtered := ihfs.Filter(fsys)

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

			filtered := ihfs.Filter(fsys, func(f *ihfs.FilterFS, o ihfs.Operation) error {
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

			filtered := ihfs.Filter(fsys,
				func(f *ihfs.FilterFS, o ihfs.Operation) error {
					if o.Subject() == "forbidden1.txt" {
						return ihfs.ErrPermission
					}
					return nil
				},
				func(f *ihfs.FilterFS, o ihfs.Operation) error {
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

			filtered := ihfs.Where(fsys, func(o ihfs.Operation) bool {
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

			filtered := ihfs.Where(fsys,
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

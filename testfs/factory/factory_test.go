package factory_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/testfs/factory"
)

var _ = Describe("Fs", func() {
	Describe("NewFs", func() {
		It("creates an Fs with default name", func() {
			f := factory.NewFs()

			Expect(f).NotTo(BeNil())
			Expect(f.Name()).To(Equal("testfs/factory"))
		})
	})

	Describe("Named", func() {
		It("sets the name and returns self", func() {
			f := factory.NewFs()

			result := f.Named("custom")

			Expect(result).To(Equal(f))
			Expect(f.Name()).To(Equal("custom"))
		})
	})

	Describe("Name", func() {
		It("returns the current name", func() {
			f := factory.NewFs().Named("my-fs")

			Expect(f.Name()).To(Equal("my-fs"))
		})
	})

	Describe("Open", func() {
		It("returns ErrNotMocked when queue is empty", func() {
			f := factory.NewFs()

			_, err := f.Open("path")

			Expect(err).To(MatchError(factory.ErrNotMocked))
		})

		It("calls the queued function", func() {
			called := false
			f := factory.NewFs().WithOpen(func(string) (ihfs.File, error) {
				called = true
				return nil, nil
			})

			_, _ = f.Open("path")

			Expect(called).To(BeTrue())
		})

		It("dequeues functions in FIFO order", func() {
			var calls []int
			f := factory.NewFs().
				WithOpen(func(string) (ihfs.File, error) { calls = append(calls, 1); return nil, nil }).
				WithOpen(func(string) (ihfs.File, error) { calls = append(calls, 2); return nil, nil })

			_, _ = f.Open("a")
			_, _ = f.Open("b")

			Expect(calls).To(Equal([]int{1, 2}))
		})

		It("returns error when queue is exhausted after dequeuing", func() {
			f := factory.NewFs().WithOpen(func(string) (ihfs.File, error) { return nil, nil })
			_, _ = f.Open("a") // consume the one mock

			_, err := f.Open("b")
			Expect(err).To(MatchError(factory.ErrNotMocked))
		})
	})

	Describe("SetOpen", func() {
		It("replaces the open queue", func() {
			var calls []int
			f := factory.NewFs().
				WithOpen(func(string) (ihfs.File, error) { calls = append(calls, 1); return nil, nil })

			f.SetOpen(func(string) (ihfs.File, error) { calls = append(calls, 2); return nil, nil })
			_, _ = f.Open("path")

			Expect(calls).To(Equal([]int{2}))
		})
	})

	Describe("Stat", func() {
		It("returns ErrNotMocked when queue is empty", func() {
			f := factory.NewFs()

			_, err := f.Stat("path")

			Expect(err).To(MatchError(factory.ErrNotMocked))
		})

		It("calls the queued function", func() {
			called := false
			f := factory.NewFs().WithStat(func(string) (ihfs.FileInfo, error) {
				called = true
				return nil, nil
			})

			_, _ = f.Stat("path")

			Expect(called).To(BeTrue())
		})
	})

	Describe("SetStat", func() {
		It("replaces the stat queue", func() {
			var calls []int
			f := factory.NewFs().
				WithStat(func(string) (ihfs.FileInfo, error) { calls = append(calls, 1); return nil, nil })

			f.SetStat(func(string) (ihfs.FileInfo, error) { calls = append(calls, 2); return nil, nil })
			_, _ = f.Stat("path")

			Expect(calls).To(Equal([]int{2}))
		})
	})
})

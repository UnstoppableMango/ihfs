package op_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/op"
)

func TestOp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Op Suite")
}

var _ = Describe("Operation types", func() {
	Describe("Open", func() {
		It("should return subject name", func() {
			o := op.Open{Name: "test.txt"}
			Expect(o.Subject()).To(Equal("test.txt"))
		})
	})

	Describe("Glob", func() {
		It("should return subject pattern", func() {
			g := op.Glob{Pattern: "*.txt"}
			Expect(g.Subject()).To(Equal("*.txt"))
		})
	})

	Describe("Lstat", func() {
		It("should return subject name", func() {
			l := op.Lstat{Name: "test.txt"}
			Expect(l.Subject()).To(Equal("test.txt"))
		})
	})

	Describe("ReadDir", func() {
		It("should return subject name", func() {
			r := op.ReadDir{Name: "test-dir"}
			Expect(r.Subject()).To(Equal("test-dir"))
		})
	})

	Describe("ReadFile", func() {
		It("should return subject name", func() {
			r := op.ReadFile{Name: "test.txt"}
			Expect(r.Subject()).To(Equal("test.txt"))
		})
	})

	Describe("ReadLink", func() {
		It("should return subject name", func() {
			r := op.ReadLink{Name: "test-link"}
			Expect(r.Subject()).To(Equal("test-link"))
		})
	})

	Describe("Stat", func() {
		It("should return subject name", func() {
			s := op.Stat{Name: "test.txt"}
			Expect(s.Subject()).To(Equal("test.txt"))
		})
	})

	Describe("WriteFile", func() {
		It("should return subject name", func() {
			w := op.WriteFile{Name: "test.txt", Data: []byte("data"), Perm: 0644}
			Expect(w.Subject()).To(Equal("test.txt"))
		})
	})

	Describe("Remove", func() {
		It("should return subject name", func() {
			r := op.Remove{Name: "test.txt"}
			Expect(r.Subject()).To(Equal("test.txt"))
		})
	})

	Describe("RemoveAll", func() {
		It("should return subject name", func() {
			r := op.RemoveAll{Name: "test-dir"}
			Expect(r.Subject()).To(Equal("test-dir"))
		})
	})
})

package op_test

import (
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/op"
)

var _ = Describe("Operation", func() {
	It("Open.Subject returns the name", func() {
		Expect(op.Open{Name: "test.txt"}.Subject()).To(Equal("test.txt"))
	})

	It("Glob.Subject returns the pattern", func() {
		Expect(op.Glob{Pattern: "*.txt"}.Subject()).To(Equal("*.txt"))
	})

	It("Lstat.Subject returns the name", func() {
		Expect(op.Lstat{Name: "link"}.Subject()).To(Equal("link"))
	})

	It("ReadDir.Subject returns the name", func() {
		Expect(op.ReadDir{Name: "dir"}.Subject()).To(Equal("dir"))
	})

	It("ReadFile.Subject returns the name", func() {
		Expect(op.ReadFile{Name: "file.txt"}.Subject()).To(Equal("file.txt"))
	})

	It("ReadLink.Subject returns the name", func() {
		Expect(op.ReadLink{Name: "symlink"}.Subject()).To(Equal("symlink"))
	})

	It("Stat.Subject returns the name", func() {
		Expect(op.Stat{Name: "info.txt"}.Subject()).To(Equal("info.txt"))
	})

	It("WriteFile.Subject returns the name", func() {
		wf := op.WriteFile{Name: "out.txt", Data: []byte("x"), Perm: 0644}
		Expect(wf.Subject()).To(Equal("out.txt"))
		Expect(wf.Data).To(Equal([]byte("x")))
		Expect(wf.Perm).To(Equal(fs.FileMode(0644)))
	})

	It("Remove.Subject returns the name", func() {
		Expect(op.Remove{Name: "old.txt"}.Subject()).To(Equal("old.txt"))
	})

	It("RemoveAll.Subject returns the name", func() {
		Expect(op.RemoveAll{Name: "dir"}.Subject()).To(Equal("dir"))
	})
})

package osfs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/osfs"
)

var _ = Describe("Fs", func() {
	It("New returns non-nil", func() {
		Expect(osfs.New()).NotTo(BeNil())
	})

	It("New can open the current directory", func() {
		f, err := osfs.New().Open(".")

		Expect(err).NotTo(HaveOccurred())
		Expect(f).NotTo(BeNil())
		Expect(f.Close()).To(Succeed())
	})

	It("Default is non-nil", func() {
		var _ ihfs.OsFS = osfs.Default
		Expect(osfs.Default).NotTo(BeNil())
	})
})

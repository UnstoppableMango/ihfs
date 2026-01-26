package osfs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/osfs"
)

func TestOsfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Osfs Suite")
}

var _ = Describe("Fs", func() {
	Describe("New", func() {
		It("should create a new OS filesystem", func() {
			fs := osfs.New()
			Expect(fs).NotTo(BeNil())
		})
	})

	Describe("Default", func() {
		It("should provide a default OS filesystem", func() {
			Expect(osfs.Default).NotTo(BeNil())
		})
	})
})

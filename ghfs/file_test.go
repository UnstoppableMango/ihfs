package ghfs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/ghfs"
)

var _ = Describe("File", func() {
	Describe("Close", func() {
		It("should return nil", func() {
			owner := &ghfs.Owner{}
			err := owner.Close()
			Expect(err).To(BeNil())

			repo := &ghfs.Repository{}
			err = repo.Close()
			Expect(err).To(BeNil())

			content := &ghfs.Content{}
			err = content.Close()
			Expect(err).To(BeNil())

			asset := &ghfs.Asset{}
			err = asset.Close()
			Expect(err).To(BeNil())

			release := &ghfs.Release{}
			err = release.Close()
			Expect(err).To(BeNil())

			branch := &ghfs.Branch{}
			err = branch.Close()
			Expect(err).To(BeNil())
		})
	})
})

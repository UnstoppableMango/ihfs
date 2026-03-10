package ghfs

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-github/v84/github"
)

var _ = Describe("release", func() {
	It("should return error when neither release ID nor tag is set", func() {
		p := Path{asset: "asset.tar.gz"}
		_, err := release(context.Background(), github.NewClient(nil), p)
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("release not specified"))
	})
})

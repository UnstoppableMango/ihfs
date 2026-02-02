package ghfs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGhfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ghfs Suite")
}

package osfs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOsfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Osfs Suite")
}

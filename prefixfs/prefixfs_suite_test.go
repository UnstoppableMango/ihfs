package prefixfs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPrefixfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prefixfs Suite")
}

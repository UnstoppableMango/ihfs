package memfs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMemfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Memfs Suite")
}

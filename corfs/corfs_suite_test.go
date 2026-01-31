package corfs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCorfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Corfs Suite")
}

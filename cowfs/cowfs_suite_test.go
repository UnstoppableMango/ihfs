package cowfs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCowfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cowfs Suite")
}

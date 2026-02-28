package testfs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTestfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Testfs Suite")
}

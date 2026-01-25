package tarfs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTarfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tarfs Suite")
}

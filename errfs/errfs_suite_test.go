package errfs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestErrfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Errfs Suite")
}

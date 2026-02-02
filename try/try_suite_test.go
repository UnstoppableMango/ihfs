package try_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Try Suite")
}

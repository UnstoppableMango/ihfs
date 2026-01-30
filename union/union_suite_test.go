package union_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUnion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Union Suite")
}

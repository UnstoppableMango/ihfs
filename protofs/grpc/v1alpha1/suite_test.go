package protofsv1alpha1_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProtofsV1alpha1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ProtofsV1alpha1 Suite")
}

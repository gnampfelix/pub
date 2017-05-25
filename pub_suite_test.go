package pub

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pub Suite")
}

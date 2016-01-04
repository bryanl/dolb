package dolbutil_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDolbutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dolbutil Suite")
}

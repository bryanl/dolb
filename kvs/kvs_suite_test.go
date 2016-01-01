package kvs_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestKvs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kvs Suite")
}

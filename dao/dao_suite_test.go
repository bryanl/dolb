package dao_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDao(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dao Suite")
}

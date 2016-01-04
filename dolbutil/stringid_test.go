package dolbutil_test

import (
	. "github.com/bryanl/dolb/dolbutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stringid", func() {
	Context("with a generated id", func() {

		var (
			id string
		)

		JustBeforeEach(func() {
			id = GenerateRandomID()
		})

		It("has a length of 64", func() {
			Ω(id).To(HaveLen(64))
		})

		It("shortens an id to 16 characters", func() {
			truncatedID := TruncateID(id)
			Ω(truncatedID).To(HaveLen(16))
		})
	})

	It("does not shorten ids of an invalid size", func() {
		id := "1234"
		truncatedID := TruncateID(id)

		Ω(len(truncatedID)).To(Equal(len(id)))
	})

	It("knows not to shorten non hex ids", func() {
		id := "this is not an id"
		Ω(IsShortID(id)).ToNot(BeTrue())
	})
})

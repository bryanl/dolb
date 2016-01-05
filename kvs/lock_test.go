package kvs_test

import (
	"time"

	. "github.com/bryanl/dolb/kvs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lock", func() {

	var (
		err error
	)

	Describe("Lock", func() {

		var (
			lock         *Lock
			lockDuration = 100 * time.Millisecond
			kvs          *MockKVS
			item         string
		)

		BeforeEach(func() {
			item = "foo"
			kvs = &MockKVS{}
			lock = NewLock(item, kvs)
		})

		AfterEach(func() {
			kvs.AssertExpectations(GinkgoT())
		})

		JustBeforeEach(func() {
			err = lock.Lock(lockDuration)
		})

		Context("locking without error", func() {

			BeforeEach(func() {
				opts := &SetOptions{TTL: lockDuration, IfNotExist: true}
				node := &Node{}
				kvs.On("Set", "/dolb/locks/foo", item, opts).Return(node, nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})

		Context("locking with a previously held lock", func() {

			BeforeEach(func() {
				node := &Node{}
				opts := &SetOptions{TTL: lockDuration, IfNotExist: true}

				errs := []error{&NodeExistError{}, &NodeExistError{}, nil}
				for _, e := range errs {
					kvs.On("Set", "/dolb/locks/foo", item, opts).Return(node, e).Once()
				}
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})

	})
})

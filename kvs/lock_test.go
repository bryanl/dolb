package kvs_test

import (
	"errors"
	"time"

	. "github.com/bryanl/dolb/kvs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lock", func() {

	var (
		err          error
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

	Describe("Lock", func() {

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

	Describe("IsLocked", func() {

		var (
			isLocked bool
			getOpts  *GetOptions
		)

		JustBeforeEach(func() {
			isLocked = lock.IsLocked()
		})

		Context("key exists", func() {

			BeforeEach(func() {
				kvs.On("Get", "/dolb/locks/foo", getOpts).Return(&Node{}, nil)
			})

			It("is locked", func() {
				Ω(isLocked).To(BeTrue())
			})
		})

		Context("key does not exist", func() {

			BeforeEach(func() {
				kvs.On("Get", "/dolb/locks/foo", getOpts).Return(nil, errors.New("hello"))
			})

			It("is unlocked", func() {
				Ω(isLocked).ToNot(BeTrue())
			})
		})

	})
})

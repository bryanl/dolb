package kvs_test

import (
	"errors"
	"fmt"
	"time"

	. "github.com/bryanl/dolb/kvs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster", func() {

	var (
		err      error
		kvs      *MockKVS
		checkTTL = time.Millisecond * 10
		cluster  *Cluster
		failErr  = errors.New("fail")
	)

	BeforeEach(func() {
		kvs = &MockKVS{}
		cluster = NewCluster(kvs, checkTTL)
	})

	Describe("RegisterAgent", func() {

		var (
			index uint64
		)

		JustBeforeEach(func() {
			index, err = cluster.RegisterAgent("agent1")
		})

		Context("with success", func() {

			BeforeEach(func() {
				opts := &SetOptions{TTL: checkTTL}
				node := &Node{ModifiedIndex: 1}
				kvs.On("Set", "/agent/leader/agent1", "agent1", opts).Return(node, nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

			It("returns the last modified index", func() {
				Ω(index).ToNot(Equal(1))
			})
		})

		Context("with kv error", func() {
			BeforeEach(func() {
				opts := &SetOptions{TTL: checkTTL}
				kvs.On("Set", "/agent/leader/agent1", "agent1", opts).Return(nil, failErr)
			})

			It("returns an error", func() {
				Ω(err).To(HaveOccurred())
			})

		})
	})

	Describe("Leader", func() {
		var (
			leader *Leader
		)

		JustBeforeEach(func() {
			leader, err = cluster.Leader()
		})

		Context("with nodes", func() {

			BeforeEach(func() {
				opts := &GetOptions{Recursive: true}

				rootNode := &Node{Nodes: Nodes{}}
				for i := uint64(1); i <= 3; i++ {
					rootNode.Nodes = append(rootNode.Nodes, &Node{CreatedIndex: i + 1, Value: fmt.Sprintf("node%d", i)})
				}

				kvs.On("Get", cluster.LeaderKey, opts).Return(rootNode, nil)
			})

			It("returns the leader", func() {
				Ω(leader).To(Equal(&Leader{Name: "node1", NodeCount: 3}))
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

		})

		Context("with no nodes", func() {
			BeforeEach(func() {
				opts := &GetOptions{Recursive: true}

				rootNode := &Node{Nodes: Nodes{}}
				kvs.On("Get", cluster.LeaderKey, opts).Return(rootNode, nil)
			})

			It("returns an error", func() {
				Ω(err).To(HaveOccurred())
			})

		})

		Context("with leader retrieval error", func() {

			BeforeEach(func() {
				opts := &GetOptions{Recursive: true}
				kvs.On("Get", cluster.LeaderKey, opts).Return(nil, failErr)

			})

			It("returns an error", func() {
				Ω(err).To(HaveOccurred())
			})

		})
	})

	Describe("Refresh", func() {

		var (
			index uint64
		)

		JustBeforeEach(func() {
			index, err = cluster.Refresh("agent1", 5)
		})

		Context("with no error", func() {

			BeforeEach(func() {
				opts := &SetOptions{
					TTL:       checkTTL,
					PrevIndex: 5,
				}
				node := &Node{ModifiedIndex: 6}
				kvs.On("Set", cluster.LeaderKey+"/agent1", "agent1", opts).Return(node, nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
			It("returns the new index", func() {
				Ω(index).To(Equal(uint64(6)))
			})
		})

		Context("with an error", func() {
			BeforeEach(func() {
				opts := &SetOptions{
					TTL:       checkTTL,
					PrevIndex: 5,
				}
				kvs.On("Set", cluster.LeaderKey+"/agent1", "agent1", opts).Return(nil, failErr)
			})

			It("returns an error", func() {
				Ω(err).To(HaveOccurred())
			})
		})
	})
})

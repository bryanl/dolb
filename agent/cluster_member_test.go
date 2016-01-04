package agent

import (
	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/kvs"
	"golang.org/x/net/context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClusterMember", func() {

	var (
		mockKVS *kvs.MockKVS
		cm      *ClusterMember
		err     error
	)

	BeforeEach(func() {
		mockKVS = &kvs.MockKVS{}
	})

	AfterEach(func() {
		mockKVS.AssertExpectations(GinkgoT())
	})

	JustBeforeEach(func() {

		cm = &ClusterMember{
			cmKVS:   kvs.NewCluster(mockKVS, 5*time.Second),
			context: context.Background(),
			logger:  logrus.WithField("testing", true),
			name:    "test",
			schedule: func(cm *ClusterMember, name string, fn scheduleFn, d time.Duration) {
				fn(cm)
			},
			poll:    poll,
			refresh: refresh,
		}
	})

	Describe("change", func() {
		It("emits the change", func() {
			newCtx, fn := context.WithCancel(cm.context)
			cm.context = newCtx

			csChan := cm.Change()

			go func() {
				cm.mu.Lock()
				defer cm.mu.Unlock()
				cm.Leader = cm.name
			}()
			cs := <-csChan
			Ω(cs.Leader).To(Equal(cm.name))
			Ω(cs.IsLeader).To(BeTrue())

			fn()
		})
	})

	Describe("stop", func() {
		Context("when not started", func() {
			It("returns an error", func() {
				err = cm.Stop()
				Ω(err).To(Equal(ErrClusterNotJoined))
			})
		})
	})

	Describe("start", func() {
		It("starts the cluster", func() {
			cm.schedule = func(*ClusterMember, string, scheduleFn, time.Duration) {
			}

			opts := &kvs.SetOptions{TTL: time.Second * 5}
			node := &kvs.Node{ModifiedIndex: 99}
			mockKVS.On("Set", "/agent/leader/test", "test", opts).Return(node, nil)

			err = cm.Start()
			Ω(err).ToNot(HaveOccurred())
			Ω(cm.started).To(BeTrue())

			err = cm.Start()
			Ω(err).To(Equal(ErrClusterJoined))

			err = cm.Stop()
			Ω(err).ToNot(HaveOccurred())
			Ω(cm.started).ToNot(BeTrue())
		})
	})

	Describe("poll", func() {
		It("updates cluster membership", func() {
			cm.started = true

			opts := &kvs.GetOptions{Recursive: true}
			node := &kvs.Node{
				Nodes: kvs.Nodes{
					{ModifiedIndex: 5, CreatedIndex: 1, Value: cm.name},
				},
			}

			mockKVS.On("Get", "/agent/leader", opts).Return(node, nil)

			poll(cm)

			Ω(cm.NodeCount).To(Equal(1))
			Ω(cm.name).To(Equal(cm.Leader))
		})
	})

	Describe("refresh", func() {
		It("refreshes the cluster membership periodically", func() {
			cm.started = true

			opts := &kvs.SetOptions{TTL: 5 * time.Second}
			node := &kvs.Node{ModifiedIndex: 99}
			mockKVS.On("Set", "/agent/leader/test", "test", opts).Return(node, nil)

			refresh(cm)

			Ω(cm.modifiedIndex).To(Equal(uint64(99)))
		})
	})

	Describe("schedule", func() {
		It("schedules operations", func() {
			cm.started = true

			ran := false

			fn := func(*ClusterMember) error {
				ran = true
				return errors.New("bye bye")
			}

			schedule(cm, "testing", fn, 100*time.Millisecond)

			Ω(ran).To(BeTrue())
		})
	})
})

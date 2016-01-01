package agent

import (
	"errors"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/kvs"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type mockEtcdClusterMember func(*ClusterMember, *myMocks)
type myMocks struct {
	KVS *kvs.MockKVS
}

func withTestClusterMember(fn mockEtcdClusterMember) {
	em := &myMocks{
		KVS: &kvs.MockKVS{},
	}

	cm := &ClusterMember{
		cmKVS:   kvs.NewCluster(em.KVS, 5*time.Second),
		context: context.Background(),
		logger:  logrus.WithField("testing", true),
		name:    "test",
		schedule: func(cm *ClusterMember, name string, fn scheduleFn, d time.Duration) {
			fn(cm)
		},
		poll:    poll,
		refresh: refresh,
	}

	fn(cm, em)
}

func TestNewClusterMember(t *testing.T) {
	name := "test"

	kv := &kvs.MockKVS{}

	c := &Config{
		KVS:     kv,
		Context: context.Background(),
	}
	cm := NewClusterMember(name, c)
	assert.NotNil(t, cm)
}

func TestClusterMember_Change(t *testing.T) {
	withTestClusterMember(func(cm *ClusterMember, em *myMocks) {
		newCtx, fn := context.WithCancel(cm.context)
		cm.context = newCtx

		csChan := cm.Change()

		go func() { cm.Leader = cm.name }()
		cs := <-csChan
		assert.Equal(t, cm.name, cs.Leader)
		assert.True(t, cs.IsLeader)

		fn()
	})
}

func TestClusterMember_Stop(t *testing.T) {
	withTestClusterMember(func(cm *ClusterMember, em *myMocks) {
		err := cm.Stop()
		assert.Equal(t, ErrClusterNotJoined, err)
	})
}

func TestClusterMember_Start(t *testing.T) {
	withTestClusterMember(func(cm *ClusterMember, em *myMocks) {
		cm.schedule = func(*ClusterMember, string, scheduleFn, time.Duration) {
			// no op
		}

		opts := &kvs.SetOptions{TTL: time.Second * 5}
		node := &kvs.Node{ModifiedIndex: 99}
		em.KVS.On("Set", "/agent/leader/test", "test", opts).Return(node, nil)

		err := cm.Start()
		assert.NoError(t, err)
		assert.True(t, cm.started)

		err = cm.Start()
		assert.Error(t, ErrClusterJoined)

		err = cm.Stop()
		assert.NoError(t, err)
		assert.False(t, cm.started)
	})
}

func Test_poll(t *testing.T) {
	withTestClusterMember(func(cm *ClusterMember, em *myMocks) {
		cm.started = true

		opts := &kvs.GetOptions{Recursive: true}
		node := &kvs.Node{
			Nodes: kvs.Nodes{
				{ModifiedIndex: 5, CreatedIndex: 1, Value: cm.name},
			},
		}

		em.KVS.On("Get", "/agent/leader", opts).Return(node, nil)

		poll(cm)

		assert.Equal(t, 1, cm.NodeCount)
		assert.Equal(t, cm.name, cm.Leader)
	})
}

func Test_refresh(t *testing.T) {
	withTestClusterMember(func(cm *ClusterMember, em *myMocks) {
		cm.started = true

		opts := &kvs.SetOptions{TTL: 5 * time.Second}
		node := &kvs.Node{ModifiedIndex: 99}
		em.KVS.On("Set", "/agent/leader/test", "test", opts).Return(node, nil)

		refresh(cm)

		assert.Equal(t, uint64(99), cm.modifiedIndex)
	})
}

func Test_schedule(t *testing.T) {
	withTestClusterMember(func(cm *ClusterMember, em *myMocks) {
		cm.started = true

		ran := false

		fn := func(*ClusterMember) error {
			ran = true
			return errors.New("bye bye")
		}

		schedule(cm, "testing", fn, 100*time.Millisecond)

		assert.True(t, ran)
	})
}

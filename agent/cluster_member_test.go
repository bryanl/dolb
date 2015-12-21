package agent

import (
	"errors"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/mocks"
	etcdclient "github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type mockEtcdClusterMember func(*ClusterMember, *etcdMocks)
type etcdMocks struct {
	KeysAPI *mocks.KeysAPI
}

func withTestClusterMember(fn mockEtcdClusterMember) {
	em := &etcdMocks{
		KeysAPI: &mocks.KeysAPI{},
	}

	cm := &ClusterMember{
		checkTTL: 5 * time.Second,
		context:  context.Background(),
		kapi:     em.KeysAPI,
		logger:   logrus.WithField("testing", true),
		name:     "test",
		schedule: func(cm *ClusterMember, name string, fn scheduleFn, d time.Duration) {
			fn(cm)
		},
		poll:    poll,
		refresh: refresh,
		root:    rootKey,
	}

	fn(cm, em)
}

func TestNewClusterMember(t *testing.T) {
	name := "test"

	ka := &mocks.KeysAPI{}

	c := &Config{
		KeysAPI: ka,
		Context: context.Background(),
	}
	cm := NewClusterMember(name, c)
	assert.NotNil(t, cm)
}

func TestClusterMember_Change(t *testing.T) {
	withTestClusterMember(func(cm *ClusterMember, em *etcdMocks) {
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
	withTestClusterMember(func(cm *ClusterMember, em *etcdMocks) {
		err := cm.Stop()
		assert.Equal(t, ErrClusterNotJoined, err)
	})
}

func TestClusterMember_Start(t *testing.T) {
	withTestClusterMember(func(cm *ClusterMember, em *etcdMocks) {
		cm.schedule = func(*ClusterMember, string, scheduleFn, time.Duration) {
			// no op
		}

		setOpts := &etcdclient.SetOptions{
			TTL: cm.checkTTL,
		}

		resp := &etcdclient.Response{
			Node: &etcdclient.Node{
				ModifiedIndex: 99,
			},
		}

		em.KeysAPI.On("Set", cm.context, cm.key(), cm.name, setOpts).Return(resp, nil)

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
	withTestClusterMember(func(cm *ClusterMember, em *etcdMocks) {
		cm.started = true

		getOpts := &etcdclient.GetOptions{Recursive: true}
		resp := &etcdclient.Response{
			Node: &etcdclient.Node{
				Nodes: etcdclient.Nodes{
					&etcdclient.Node{
						ModifiedIndex: 5,
						CreatedIndex:  1,
						Value:         cm.name,
					},
				},
			},
		}
		em.KeysAPI.On("Get", cm.context, cm.root, getOpts).Return(resp, nil)

		poll(cm)

		assert.Equal(t, 1, cm.NodeCount)
		assert.Equal(t, cm.name, cm.Leader)
	})
}

func Test_refresh(t *testing.T) {
	withTestClusterMember(func(cm *ClusterMember, em *etcdMocks) {
		cm.started = true

		setOpts := &etcdclient.SetOptions{
			TTL: cm.checkTTL,
		}

		resp := &etcdclient.Response{
			Node: &etcdclient.Node{
				ModifiedIndex: 99,
			},
		}

		em.KeysAPI.On("Set", cm.context, cm.key(), cm.name, setOpts).Return(resp, nil)

		refresh(cm)

		assert.Equal(t, uint64(99), cm.modifiedIndex)
	})
}

func Test_schedule(t *testing.T) {
	withTestClusterMember(func(cm *ClusterMember, em *etcdMocks) {
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

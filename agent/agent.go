package agent

import (
	"errors"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	etcdclient "github.com/coreos/etcd/client"
)

var (
	// checkTTL is the time to live for cluster member keys
	checkTTL = 10 * time.Second

	// rootKey is the root key for cluster memberships. Members will create their keys
	// below this one.
	rootKey = "/agent/leader/"

	// ErrClusterNotJoined is returned when this agent has not joined a cluster.
	ErrClusterNotJoined = errors.New("agent has not joined the cluster")

	// ErrClusterJoined is returned when this agent has already joined a cluster.
	ErrClusterJoined = errors.New("agent has already joined the cluster")
)

// ClusterStatus is the status of the cluster.
type ClusterStatus struct {
	Leader    string `json:"leader"`
	NodeCount int    `json:"node_count"`
}

// ClusterMember is an agent cluster membership.
type ClusterMember struct {
	checkTTL time.Duration
	context  context.Context
	kapi     etcdclient.KeysAPI
	name     string
	root     string

	Leader    string
	NodeCount int

	started       bool
	modifiedIndex uint64

	schedule func(*ClusterMember, string, scheduleFn, time.Duration)
	poll     func(el *ClusterMember) error
	refresh  func(el *ClusterMember) error

	logger *logrus.Entry
}

// NewClusterMember builds a ClusterMember.
func NewClusterMember(ctx context.Context, name string, kapi etcdclient.KeysAPI) *ClusterMember {
	return &ClusterMember{
		checkTTL: checkTTL,
		context:  ctx,
		kapi:     kapi,
		logger: logrus.WithFields(logrus.Fields{
			"member-name": name,
		}),
		name:    name,
		refresh: refresh,

		schedule: schedule,
		poll:     poll,
		root:     rootKey,
	}
}

// Change creates a channel that outputs the current cluster leader.
func (cm *ClusterMember) Change() chan ClusterStatus {
	t := time.NewTicker(time.Millisecond * 250)
	out := make(chan ClusterStatus, 1)

	leader := cm.Leader

	go func() {
		for {
			select {
			case <-t.C:
				if cm.Leader != "" && leader != cm.Leader {
					cs := ClusterStatus{
						Leader:    cm.Leader,
						NodeCount: cm.NodeCount,
					}
					out <- cs
				}
			case <-cm.context.Done():
				break
			}
		}
	}()

	return out
}

func (cm *ClusterMember) key() string {
	return cm.root + cm.name
}

// Start starts a cluster membership process.
func (cm *ClusterMember) Start() error {
	if cm.started {
		return ErrClusterJoined
	}

	cm.started = true

	opts := &etcdclient.SetOptions{
		TTL: cm.checkTTL,
	}

	repo, err := cm.kapi.Set(cm.context, cm.key(), cm.name, opts)
	if err != nil {
		logrus.WithError(err).Error("cannot set initial value")
		return err
	}

	cm.modifiedIndex = repo.Node.ModifiedIndex

	go cm.schedule(cm, "poll", poll, time.Second)
	go cm.schedule(cm, "refresh", refresh, cm.checkTTL/2)

	return nil
}

// Stop stops a cluster membership process.
func (cm *ClusterMember) Stop() error {
	if !cm.started {
		return ErrClusterNotJoined
	}

	cm.started = false
	return nil
}

type scheduleFn func(*ClusterMember) error

func schedule(cm *ClusterMember, name string, fn scheduleFn, timeout time.Duration) {
	logger := cm.logger.WithField("cluster-action", name)

	t := time.NewTicker(timeout)
	quit := make(chan struct{})

	for {
		if !cm.started {
			t.Stop()
			close(quit)
			break
		}

		select {
		case <-t.C:
			err := fn(cm)
			if err != nil {
				logger.WithError(err).Error("could not run scheduled item")
				t.Stop()
				close(quit)
			}
		case <-quit:
			logger.Info("shutting down")
			return
		}
	}
}

func poll(cm *ClusterMember) error {
	opts := &etcdclient.GetOptions{
		Recursive: true,
	}

	resp, err := cm.kapi.Get(cm.context, cm.root, opts)
	if err != nil {
		return err
	}

	min := resp.Node.Nodes[0].ModifiedIndex
	var leaderNode etcdclient.Node
	for _, n := range resp.Node.Nodes {
		if n.CreatedIndex < min {
			min = n.CreatedIndex
			leaderNode = *n
		}
	}

	if leader := leaderNode.Value; leader != "" && leader != cm.Leader {
		cm.logger.WithFields(logrus.Fields{
			"leader": leaderNode.Value,
		}).Info("updated leader")

		cm.Leader = leaderNode.Value
	}

	if l := len(resp.Node.Nodes); l != cm.NodeCount {
		cm.NodeCount = l
		cm.logger.WithFields(logrus.Fields{
			"node-count": l,
		}).Info("updated node count")
	}

	return nil
}

func refresh(cm *ClusterMember) error {
	opts := &etcdclient.SetOptions{
		TTL:       cm.checkTTL,
		PrevIndex: cm.modifiedIndex,
	}

	resp, err := cm.kapi.Set(cm.context, cm.key(), cm.name, opts)
	if err != nil {
		return err
	}

	cm.modifiedIndex = resp.Node.ModifiedIndex

	return nil
}

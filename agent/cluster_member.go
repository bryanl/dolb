package agent

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/kvs"
)

var (
	// checkTTL is the time to live for cluster member keys
	checkTTL = 10 * time.Second

	// ErrClusterNotJoined is returned when this agent has not joined a cluster.
	ErrClusterNotJoined = errors.New("agent has not joined the cluster")

	// ErrClusterJoined is returned when this agent has already joined a cluster.
	ErrClusterJoined = errors.New("agent has already joined the cluster")
)

// RegisterError is a cluster registration error.
type RegisterError struct {
	name string
	err  error
}

func (re *RegisterError) Error() string {
	return fmt.Sprintf("unable to register agent %q: %v", re.name, re.err)
}

// ClusterStatus is the status of the cluster.
type ClusterStatus struct {
	FloatingIP string `json:"floating_ip"`
	Leader     string `json:"leader"`
	IsLeader   bool   `json:"is_leader"`
	NodeCount  int    `json:"node_count"`
}

// ClusterMember is an agent cluster membership.
type ClusterMember struct {
	cmKVS   *kvs.Cluster
	context context.Context
	name    string
	root    string

	Leader    string
	NodeCount int

	started       bool
	modifiedIndex uint64

	schedule func(*ClusterMember, string, scheduleFn, time.Duration)
	poll     func(el *ClusterMember) error
	refresh  func(el *ClusterMember) error

	logger *logrus.Entry
	mu     sync.Mutex
}

// NewClusterMember builds a ClusterMember.
func NewClusterMember(name string, config *Config) *ClusterMember {
	return &ClusterMember{
		cmKVS:   kvs.NewCluster(config.KVS, checkTTL),
		context: config.Context,
		logger: logrus.WithFields(logrus.Fields{
			"member-name": name,
		}),
		name:    name,
		refresh: refresh,

		schedule: schedule,
		poll:     poll,
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
				cm.mu.Lock()
				currentLeader := cm.Leader
				cm.mu.Unlock()

				if leader != currentLeader {
					cs := ClusterStatus{
						IsLeader:  cm.isLeader(),
						Leader:    currentLeader,
						NodeCount: cm.NodeCount,
					}
					leader = currentLeader
					out <- cs
				}
			case <-cm.context.Done():
				break
			}
		}
	}()

	return out
}

func (cm *ClusterMember) isLeader() bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.name == cm.Leader
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

	mi, err := cm.cmKVS.RegisterAgent(cm.name)
	if err != nil {
		return &RegisterError{err: err, name: cm.name}
	}

	cm.modifiedIndex = mi

	go cm.schedule(cm, "poll", poll, time.Second)
	go cm.schedule(cm, "refresh", refresh, cm.cmKVS.CheckTTL/2)

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
	leader, err := cm.cmKVS.Leader()
	if err != nil {
		return err
	}

	logMsg := cm.logger
	shouldLog := false

	cm.mu.Lock()
	if l := leader.Name; cm.Leader != l {
		logMsg = logMsg.WithField("leader", l)
		cm.Leader = l
		shouldLog = true
	}
	cm.mu.Unlock()

	if nc := leader.NodeCount; cm.NodeCount != nc {
		logMsg = logMsg.WithField("node-count", nc)
		cm.NodeCount = nc
		shouldLog = true
	}

	if shouldLog {
		logMsg.Info("cluster updated")
	}

	return nil
}

func refresh(cm *ClusterMember) error {
	mi, err := cm.cmKVS.Refresh(cm.name, cm.modifiedIndex)
	if err != nil {
		return err
	}

	cm.modifiedIndex = mi

	return nil
}

package kvs

import (
	"fmt"
	"time"
)

// CmKVS is a cluster management kvs.
type Cluster struct {
	KVS

	CheckTTL  time.Duration
	LeaderKey string
}

// NewCmKVS builds a CmKVS instance.
func NewCluster(backend KVS, checkTTL time.Duration) *Cluster {
	return &Cluster{
		KVS:       backend,
		CheckTTL:  checkTTL,
		LeaderKey: "/agent/leader",
	}
}

// RegisterAgent register an agent in the kvs.
func (ckvs *Cluster) RegisterAgent(name string) (uint64, error) {
	opts := &SetOptions{
		TTL: ckvs.CheckTTL,
	}

	key := fmt.Sprintf("%s/%s", ckvs.LeaderKey, name)
	node, err := ckvs.Set(key, name, opts)
	if err != nil {
		return 0, err
	}

	return node.ModifiedIndex, nil
}

// Leader is a cluster leader.
type Leader struct {
	Name      string
	NodeCount int
}

// Leader returns a leader based on a key's CreatedIndex.
func (ckvs *Cluster) Leader() (*Leader, error) {
	opts := &GetOptions{Recursive: true}
	rootNode, err := ckvs.Get(ckvs.LeaderKey, opts)
	if err != nil {
		return nil, err
	}

	if len(rootNode.Nodes) == 0 {
		return nil, fmt.Errorf("%q has no members", ckvs.LeaderKey)
	}

	min := rootNode.Nodes[0].CreatedIndex
	leader := rootNode.Nodes[0].Value

	for _, n := range rootNode.Nodes {
		if n.CreatedIndex <= min {
			min = n.CreatedIndex
			leader = n.Value
		}
	}

	return &Leader{
		Name:      leader,
		NodeCount: len(rootNode.Nodes),
	}, nil
}

// Refresh refreshes a key with a new TTL.
func (ckvs *Cluster) Refresh(name string, lastIndex uint64) (uint64, error) {
	opts := &SetOptions{
		TTL:       ckvs.CheckTTL,
		PrevIndex: lastIndex,
	}

	key := fmt.Sprintf("%s/%s", ckvs.LeaderKey, name)
	node, err := ckvs.Set(key, name, opts)
	if err != nil {
		return 0, err
	}

	return node.ModifiedIndex, nil
}

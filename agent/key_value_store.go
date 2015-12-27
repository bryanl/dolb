package agent

import (
	"fmt"
	"time"

	etcdclient "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// MkdirError is an a kvs mkdir error.
type MkdirError struct {
	dir string
	err error
}

func (mde *MkdirError) Error() string {
	return fmt.Sprintf("could not create %q directory: %v", mde.dir, mde.err)
}

// KVError is a general kvs error.
type KVError struct {
	key string
	err error
}

func (ke *KVError) Error() string {
	return fmt.Sprintf("could not set %q: %v", ke.key, ke.err)
}

// KVDeleteError is a error while deleting an entry from the kvs.
type KVDeleteError struct {
	key string
	err error
}

func (kde *KVDeleteError) Error() string {
	return fmt.Sprintf("could not delete %q: %v", kde.key, kde.err)
}

// Node represents an entry in the kvs.
type Node struct {
	CreatedIndex  uint64
	Dir           bool
	Expiration    *time.Time
	Key           string
	ModifiedIndex uint64
	Nodes         Nodes
	Value         string
}

// Nodes is a slice of Node pointers.
type Nodes []*Node

// KVS is an interface for operations on a kvs.
type KVS interface {
	Delete(key string) error
	Get(key string, options *GetOptions) (*Node, error)
	Mkdir(dir string) error
	Rmdir(dir string) error
	Set(key, value string, options *SetOptions) (*Node, error)
}

// GetOptions are options for get operations.
type GetOptions struct {
	Recursive bool
}

// SetOptions are options for set operations.
type SetOptions struct {
	TTL       time.Duration
	PrevIndex uint64
}

// EtcdKVS is a kvs based on etcd.
type EtcdKVS struct {
	ctx   context.Context
	ksapi etcdclient.KeysAPI
}

var _ KVS = &EtcdKVS{}

// NewEtcdKVS builds a EtcdKVS instance.
func NewEtcdKVS(ctx context.Context, ksapi etcdclient.KeysAPI) *EtcdKVS {
	return &EtcdKVS{
		ctx:   ctx,
		ksapi: ksapi,
	}
}

// Mkdir makes a directory in the kvs.
func (ekvs *EtcdKVS) Mkdir(dir string) error {
	opts := &etcdclient.SetOptions{
		Dir: true,
	}

	_, err := ekvs.ksapi.Set(ekvs.ctx, dir, "", opts)
	if err != nil {
		return &MkdirError{
			dir: dir,
			err: err,
		}
	}

	return nil
}

// Set creates or updates a key in the kvs.
func (ekvs *EtcdKVS) Set(key, value string, options *SetOptions) (*Node, error) {
	if options == nil {
		options = &SetOptions{}
	}

	opts := &etcdclient.SetOptions{
		TTL:       options.TTL,
		PrevIndex: options.PrevIndex,
	}

	resp, err := ekvs.ksapi.Set(ekvs.ctx, key, value, opts)
	if err != nil {
		return nil, &KVError{
			key: key,
			err: err,
		}
	}

	n := ekvs.convertNode(resp.Node)

	return n, nil
}

// Get retrieves a key from the kvs.
func (ekvs *EtcdKVS) Get(key string, options *GetOptions) (*Node, error) {
	if options == nil {
		options = &GetOptions{}
	}

	opts := &etcdclient.GetOptions{
		Recursive: options.Recursive,
	}

	resp, err := ekvs.ksapi.Get(ekvs.ctx, key, opts)
	if err != nil {
		return nil, &KVError{
			key: key,
			err: err,
		}
	}
	n := ekvs.convertNode(resp.Node)

	for _, en := range resp.Node.Nodes {
		n.Nodes = append(n.Nodes, ekvs.convertNode(en))
	}

	return n, nil
}

// convertNode converts an etcd client Node to a Node.
func (ekvs *EtcdKVS) convertNode(in *etcdclient.Node) *Node {
	return &Node{
		CreatedIndex:  in.CreatedIndex,
		Dir:           in.Dir,
		Expiration:    in.Expiration,
		Key:           in.Key,
		ModifiedIndex: in.ModifiedIndex,
		Nodes:         Nodes{},
		Value:         in.Value,
	}
}

// Rmdir removes a directory from the kvs.
func (ekvs *EtcdKVS) Rmdir(dir string) error {
	opts := &etcdclient.DeleteOptions{
		Dir: true,
	}

	_, err := ekvs.ksapi.Delete(ekvs.ctx, dir, opts)
	if err != nil {
		return &KVDeleteError{
			key: dir,
			err: err,
		}
	}

	return nil
}

// Delete deletes a key from the kvs.
func (ekvs *EtcdKVS) Delete(key string) error {
	opts := &etcdclient.DeleteOptions{}

	_, err := ekvs.ksapi.Delete(ekvs.ctx, key, opts)
	if err != nil {
		return &KVDeleteError{
			key: key,
			err: err,
		}
	}

	return nil
}

// HaproxyKVS is a haproxy management kvs.
type HaproxyKVS struct {
	KVS
}

// NewHaproxyKVS builds a HaproxyKVS instance.
func NewHaproxyKVS(backend KVS) *HaproxyKVS {
	return &HaproxyKVS{
		KVS: backend,
	}
}

// Init initializes a kvs for haproxy configuration management.
func (hkvs *HaproxyKVS) Init() error {
	err := hkvs.Mkdir("/haproxy-discover/services")
	if err != nil {
		return err
	}

	err = hkvs.Mkdir("/haproxy-discover/tcp-services")
	if err != nil {
		return err
	}

	return nil
}

// Domain creates an endpoint based on a domain name.
func (hkvs *HaproxyKVS) Domain(app, domain string) error {
	key := fmt.Sprintf("/haproxy-discover/services/%s/domain", app)
	_, err := hkvs.Set(key, domain, nil)
	return err
}

// URLReg creates an endpoint based on a regular expression.
func (hkvs *HaproxyKVS) URLReg(app, reg string) error {
	key := fmt.Sprintf("/haproxy-discover/services/%s/url_reg", app)
	_, err := hkvs.Set(key, reg, nil)
	return err
}

// Upstream sets a new upstream node.
func (hkvs *HaproxyKVS) Upstream(app, service, address string) error {
	key := fmt.Sprintf("/haproxy-discover/services/%s/upstreams/%s", app, service)
	_, err := hkvs.Set(key, address, nil)
	return err
}

// FipKVS is a floating ip management kvs.
type FipKVS struct {
	KVS
}

// NewFipKVS builds a FipKVS instance.
func NewFipKVS(backend KVS) *FipKVS {
	return &FipKVS{
		KVS: backend,
	}
}

// CmKVS is a cluster management kvs.
type CmKVS struct {
	KVS

	CheckTTL  time.Duration
	LeaderKey string
}

// NewCmKVS builds a CmKVS instance.
func NewCmKVS(backend KVS, checkTTL time.Duration) *CmKVS {
	return &CmKVS{
		KVS:       backend,
		CheckTTL:  checkTTL,
		LeaderKey: "/agent/leader",
	}
}

// RegisterAgent register an agent in the kvs.
func (ckvs *CmKVS) RegisterAgent(name string) (uint64, error) {
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

// Leader is a leader.
type Leader struct {
	Name      string
	NodeCount int
}

// Leader returns a leader based on a key's CreatedIndex.
func (ckvs *CmKVS) Leader() (*Leader, error) {
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
func (ckvs *CmKVS) Refresh(name string, lastIndex uint64) (uint64, error) {
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

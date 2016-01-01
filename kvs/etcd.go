package kvs

import (
	etcdclient "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// EtcdKVS is a kvs based on etcd.
type Etcd struct {
	ctx   context.Context
	ksapi etcdclient.KeysAPI
}

var _ KVS = &Etcd{}

// NewEtcdKVS builds a EtcdKVS instance.
func NewEtcd(ctx context.Context, ksapi etcdclient.KeysAPI) *Etcd {
	return &Etcd{
		ctx:   ctx,
		ksapi: ksapi,
	}
}

// Mkdir makes a directory in the kvs.
func (ekvs *Etcd) Mkdir(dir string) error {
	opts := &etcdclient.SetOptions{
		Dir: true,
	}

	_, err := ekvs.ksapi.Set(ekvs.ctx, dir, "", opts)
	if err != nil {
		return &MkdirError{
			Dir: dir,
			Err: err,
		}
	}

	return nil
}

// Set creates or updates a key in the kvs.
func (ekvs *Etcd) Set(key, value string, options *SetOptions) (*Node, error) {
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
			Key: key,
			Err: err,
		}
	}

	n := ekvs.convertNode(resp.Node)

	return n, nil
}

// Get retrieves a key from the kvs.
func (ekvs *Etcd) Get(key string, options *GetOptions) (*Node, error) {
	if options == nil {
		options = &GetOptions{}
	}

	opts := &etcdclient.GetOptions{
		Recursive: options.Recursive,
	}

	resp, err := ekvs.ksapi.Get(ekvs.ctx, key, opts)
	if err != nil {
		return nil, &KVError{
			Key: key,
			Err: err,
		}
	}
	n := ekvs.convertNode(resp.Node)

	for _, en := range resp.Node.Nodes {
		n.Nodes = append(n.Nodes, ekvs.convertNode(en))
	}

	return n, nil
}

// convertNode converts an etcd client Node to a Node.
func (ekvs *Etcd) convertNode(in *etcdclient.Node) *Node {
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
func (ekvs *Etcd) Rmdir(dir string) error {
	opts := &etcdclient.DeleteOptions{
		Dir: true,
	}

	_, err := ekvs.ksapi.Delete(ekvs.ctx, dir, opts)
	if err != nil {
		return &KVDeleteError{
			Key: dir,
			Err: err,
		}
	}

	return nil
}

// Delete deletes a key from the kvs.
func (ekvs *Etcd) Delete(key string) error {
	opts := &etcdclient.DeleteOptions{}

	_, err := ekvs.ksapi.Delete(ekvs.ctx, key, opts)
	if err != nil {
		return &KVDeleteError{
			Key: key,
			Err: err,
		}
	}

	return nil
}

package kvs

import "time"

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

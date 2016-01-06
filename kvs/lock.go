package kvs

import "time"

const (
	// baseLockDir is the base directory for locks.
	baseLockDir = "/dolb/locks"
)

var (
	// lockRetryTimeout is the timeout for when a lock is active. It
	// will wait for lockRetryTimeout and try again.
	lockRetryTimeout = time.Millisecond * 100
)

// KLock is an interface for proving locks.
type KLock interface {
	Lock(time.Duration) error
	Unlock() error
}

// KLock provides a distributed lock service based on a key value store.
type Lock struct {
	KVS
	BaseDir string
	Item    string
}

var _ KLock = &Lock{}

// NewLock creates a Lock
func NewLock(item string, kvs KVS) *Lock {
	return &Lock{
		BaseDir: baseLockDir,
		Item:    item,
		KVS:     kvs,
	}
}

// Lock locks the lock by setting a key. If the key exists, it'll retry
// 100 milliseconds later (by default)
func (el *Lock) Lock(d time.Duration) error {
	for {
		opts := &SetOptions{TTL: d, IfNotExist: true}
		_, err := el.Set(el.key(), el.Item, opts)
		if eerr, ok := err.(*NodeExistError); !ok {
			return eerr
		}

		if err == nil {
			break
		}

		time.Sleep(lockRetryTimeout)
	}

	return nil
}

func (el *Lock) IsLocked() bool {
	_, err := el.Get(el.key(), nil)
	return err == nil
}

// Unlock the lock by deleting the key in kvs.
func (el *Lock) Unlock() error {
	return el.Delete(el.key())
}

func (el *Lock) key() string {
	return el.BaseDir + "/" + el.Item
}

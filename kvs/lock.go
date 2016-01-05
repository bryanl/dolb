package kvs

import "time"

const (
	baseLockDir = "/dolb/locks"
)

var (
	lockRetryTimeout = time.Millisecond * 100
)

type KLock interface {
	Lock(time.Duration) error
	Unlock() error
}

type Lock struct {
	KVS
	BaseDir string
	Item    string
}

var _ KLock = &Lock{}

func NewLock(item string, kvs KVS) *Lock {
	return &Lock{
		BaseDir: baseLockDir,
		Item:    item,
		KVS:     kvs,
	}
}

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

func (el *Lock) Unlock() error {
	return el.Delete(el.key())
}

func (el *Lock) key() string {
	return el.BaseDir + "/" + el.Item
}

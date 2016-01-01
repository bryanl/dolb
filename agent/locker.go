package agent

import (
	"sync"
	"time"

	"github.com/bryanl/dolb/kvs"
	etcdclient "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// Locker locks and blocks until it is unlocked.
type Locker interface {
	Lock() error
	Unlock() error
}

type etcdLocker struct {
	context context.Context
	key     string
	who     string
	kv      kvs.KVS
}

func newEtcdLocker(context context.Context, key, who string, kv kvs.KVS) *etcdLocker {
	return &etcdLocker{
		context: context,
		key:     key,
		who:     who,
		kv:      kv,
	}
}

var _ Locker = &etcdLocker{}

func (el *etcdLocker) Lock() error {
	for {
		_, err := el.kv.Get(el.lockKey(), nil)
		if err != nil {
			if eerr, ok := err.(etcdclient.Error); ok && eerr.Code == etcdclient.ErrorCodeNodeExist {
				break
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	_, err := el.kv.Set(el.lockKey(), el.who, nil)
	return err
}

func (el *etcdLocker) Unlock() error {
	return el.kv.Delete(el.lockKey())
}

func (el *etcdLocker) lockKey() string {
	return el.key + ".lock"
}

type memLocker struct {
	mu sync.Mutex
}

var _ Locker = &memLocker{}

func (ml *memLocker) Lock() error {
	ml.mu.Lock()
	return nil
}

func (ml *memLocker) Unlock() error {
	ml.mu.Unlock()
	return nil
}

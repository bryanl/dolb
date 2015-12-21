package agent

import (
	"sync"
	"time"

	etcdclient "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type Locker interface {
	Lock() error
	Unlock() error
}

type etcdLocker struct {
	context context.Context
	key     string
	who     string
	kapi    etcdclient.KeysAPI
}

func newEtcdLocker(context context.Context, key, who string, kapi etcdclient.KeysAPI) *etcdLocker {
	return &etcdLocker{
		context: context,
		key:     key,
		who:     who,
		kapi:    kapi,
	}
}

var _ Locker = &etcdLocker{}

func (el *etcdLocker) Lock() error {
	for {
		opts := &etcdclient.GetOptions{}
		_, err := el.kapi.Get(el.context, el.lockKey(), opts)
		if err != nil {
			if eerr, ok := err.(etcdclient.Error); ok && eerr.Code == etcdclient.ErrorCodeNodeExist {
				break
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	setOpts := &etcdclient.SetOptions{}
	_, err := el.kapi.Set(el.context, el.lockKey(), el.who, setOpts)
	return err
}

func (el *etcdLocker) Unlock() error {
	opts := &etcdclient.DeleteOptions{}
	_, err := el.kapi.Delete(el.context, el.lockKey(), opts)
	return err
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

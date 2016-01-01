package agent

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/bryanl/dolb/kvs"
	etcdclient "github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type keysAPIfn func(el *etcdLocker, kv *kvs.MockKVS)

func withKeysAPI(fn keysAPIfn) {
	m := &kvs.MockKVS{}
	el := newEtcdLocker(context.Background(), "/foo", "user-a", m)

	fn(el, m)
}

func Test_etcdLocker(t *testing.T) {
	withKeysAPI(func(el *etcdLocker, kv *kvs.MockKVS) {
		getErr := etcdclient.Error{
			Code: etcdclient.ErrorCodeNodeExist,
		}

		node := &kvs.Node{}

		kv.On("Get", "/foo.lock", mock.Anything).Return(node, nil).Once()
		kv.On("Get", "/foo.lock", mock.Anything).Return(nil, getErr).Once()
		kv.On("Set", "/foo.lock", el.who, mock.Anything).Return(&kvs.Node{}, nil)
		kv.On("Delete", "/foo.lock").Return(nil)

		err := el.Lock()
		assert.NoError(t, err)
		err = el.Unlock()
		assert.NoError(t, err)
	})
}

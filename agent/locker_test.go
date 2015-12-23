package agent

import (
	"testing"

	"golang.org/x/net/context"

	etcdclient "github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type keysAPIfn func(el *etcdLocker, kvs *mockKVS)

func withKeysAPI(fn keysAPIfn) {
	m := &mockKVS{}
	el := newEtcdLocker(context.Background(), "/foo", "user-a", m)

	fn(el, m)
}

func Test_etcdLocker(t *testing.T) {
	withKeysAPI(func(el *etcdLocker, kvs *mockKVS) {
		getErr := etcdclient.Error{
			Code: etcdclient.ErrorCodeNodeExist,
		}

		node := &Node{}

		kvs.On("Get", "/foo.lock", mock.Anything).Return(node, nil).Once()
		kvs.On("Get", "/foo.lock", mock.Anything).Return(nil, getErr).Once()
		kvs.On("Set", "/foo.lock", el.who, mock.Anything).Return(&Node{}, nil)
		kvs.On("Delete", "/foo.lock").Return(nil)

		err := el.Lock()
		assert.NoError(t, err)
		err = el.Unlock()
		assert.NoError(t, err)
	})
}

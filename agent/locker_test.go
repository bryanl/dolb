package agent

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/bryanl/dolb/mocks"
	etcdclient "github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type keysAPIfn func(el *etcdLocker, ka *mocks.KeysAPI)

func withKeysAPI(fn keysAPIfn) {
	ka := &mocks.KeysAPI{}
	el := newEtcdLocker(context.Background(), "/foo", "user-a", ka)

	fn(el, ka)
}

func Test_etcdLocker(t *testing.T) {
	withKeysAPI(func(el *etcdLocker, ka *mocks.KeysAPI) {
		getErr := etcdclient.Error{
			Code: etcdclient.ErrorCodeNodeExist,
		}
		ka.On("Get", el.context, "/foo.lock", mock.Anything).Return(nil, nil).Once()
		ka.On("Get", el.context, "/foo.lock", mock.Anything).Return(nil, getErr)
		ka.On("Set", el.context, "/foo.lock", el.who, mock.Anything).Return(nil, nil)
		ka.On("Delete", el.context, "/foo.lock", mock.Anything).Return(nil, nil)

		err := el.Lock()
		assert.NoError(t, err)
		err = el.Unlock()
		assert.NoError(t, err)
	})
}

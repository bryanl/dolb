package agent

import (
	"errors"
	"testing"

	"golang.org/x/net/context"

	"github.com/bryanl/dolb/mocks"
	"github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
)

func Test_mkdirError(t *testing.T) {
	err := mkdirError{
		dir: "/foo",
		err: errors.New("too much foo"),
	}

	assert.Equal(t, `could not create "/foo" directory: too much foo`, err.Error())
}

func Test_kvError(t *testing.T) {
	err := kvError{
		key: "/foo",
		err: errors.New("too much foo"),
	}

	assert.Equal(t, `could not set "/foo": too much foo`, err.Error())
}

func Test_kvDeleteError(t *testing.T) {
	err := kvDeleteError{
		key: "/foo",
		err: errors.New("too much foo"),
	}

	assert.Equal(t, `could not delete "/foo": too much foo`, err.Error())
}

func Test_newEtcdKVS(t *testing.T) {
	ctx := context.Background()
	ksapi := &mocks.KeysAPI{}
	kvs := newEtcdKVS(ctx, ksapi)
	assert.NotNil(t, kvs)
}

func Test_etcdKVS_mkdir(t *testing.T) {
	ctx := context.Background()
	ksapi := &mocks.KeysAPI{}
	kvs := newEtcdKVS(ctx, ksapi)

	opts := &client.SetOptions{
		Dir: true,
	}

	dir := "/foo"

	ksapi.On("Set", ctx, dir, "", opts).Return(nil, nil).Once()
	err := kvs.mkdir(dir)
	assert.NoError(t, err)

	ksapi.On("Set", ctx, dir, "", opts).Return(nil, errors.New("error")).Once()
	err = kvs.mkdir(dir)
	assert.Error(t, err)
}

func Test_etcdKVS_set(t *testing.T) {
	ctx := context.Background()
	ksapi := &mocks.KeysAPI{}
	kvs := newEtcdKVS(ctx, ksapi)

	opts := &client.SetOptions{}

	k := "/foo"
	v := "bar"

	ksapi.On("Set", ctx, k, v, opts).Return(nil, nil).Once()
	err := kvs.set(k, v)
	assert.NoError(t, err)

	ksapi.On("Set", ctx, k, v, opts).Return(nil, errors.New("error")).Once()
	err = kvs.set(k, v)
	assert.Error(t, err)
}

func Test_etcdKVS_rmdir(t *testing.T) {
	ctx := context.Background()
	ksapi := &mocks.KeysAPI{}
	kvs := newEtcdKVS(ctx, ksapi)

	opts := &client.DeleteOptions{
		Dir: true,
	}

	dir := "/foo"

	ksapi.On("Delete", ctx, dir, opts).Return(nil, nil).Once()
	err := kvs.rmdir(dir)
	assert.NoError(t, err)

	ksapi.On("Delete", ctx, dir, opts).Return(nil, errors.New("error")).Once()
	err = kvs.rmdir(dir)
	assert.Error(t, err)
}

func Test_etcdKVS_delete(t *testing.T) {
	ctx := context.Background()
	ksapi := &mocks.KeysAPI{}
	kvs := newEtcdKVS(ctx, ksapi)

	opts := &client.DeleteOptions{}

	key := "/foo"

	ksapi.On("Delete", ctx, key, opts).Return(nil, nil).Once()
	err := kvs.delete(key)
	assert.NoError(t, err)

	ksapi.On("Delete", ctx, key, opts).Return(nil, errors.New("error")).Once()
	err = kvs.delete(key)
	assert.Error(t, err)
}

func Test_etcdKVS_get(t *testing.T) {
	ctx := context.Background()
	ksapi := &mocks.KeysAPI{}
	kvs := newEtcdKVS(ctx, ksapi)

	opts := &client.GetOptions{}

	key := "/foo"

	resp := &client.Response{
		Node: &client.Node{
			Value: "bar",
		},
	}

	ksapi.On("Get", ctx, key, opts).Return(resp, nil).Once()
	bar, err := kvs.get(key)
	assert.NoError(t, err)
	assert.Equal(t, "bar", bar)

	ksapi.On("Get", ctx, key, opts).Return(nil, errors.New("error")).Once()
	_, err = kvs.get(key)
	assert.Error(t, err)
}

func Test_newHaproxyKVS(t *testing.T) {
	mk := &mockKVS{}
	hk := newHaproxyKVS(mk)

	assert.NotNil(t, hk)
}

func Test_haproxyKVS_init(t *testing.T) {
	mk := &mockKVS{}
	hk := newHaproxyKVS(mk)

	mk.On("mkdir", "/haproxy-discover/services").Return(nil).Once()
	mk.On("mkdir", "/haproxy-discover/tcp-services").Return(nil).Once()

	err := hk.Init()
	assert.NoError(t, err)
}

func Test_haproxyKVS_domain(t *testing.T) {
	mk := &mockKVS{}
	hk := newHaproxyKVS(mk)

	mk.On("set", "/haproxy-discover/services/app/domain", "domain").Return(nil).Once()
	err := hk.domain("app", "domain")
	assert.NoError(t, err)
}

func Test_haproxyKVS_urlReg(t *testing.T) {
	mk := &mockKVS{}
	hk := newHaproxyKVS(mk)

	mk.On("set", "/haproxy-discover/services/app/url_reg", "regex").Return(nil).Once()
	err := hk.urlReg("app", "regex")
	assert.NoError(t, err)
}

func Test_haproxyKVS_upstream(t *testing.T) {
	mk := &mockKVS{}
	hk := newHaproxyKVS(mk)

	mk.On("set", "/haproxy-discover/services/app/upstreams/nodeA", "127.0.0.1:80").Return(nil).Once()
	err := hk.upstream("app", "nodeA", "127.0.0.1:80")
	assert.NoError(t, err)
}

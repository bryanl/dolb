package agent

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/bryanl/dolb/mocks"
	"github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_MkdirError(t *testing.T) {
	err := MkdirError{
		dir: "/foo",
		err: errors.New("too much foo"),
	}

	assert.Equal(t, `could not create "/foo" directory: too much foo`, err.Error())
}

func Test_KVError(t *testing.T) {
	err := KVError{
		key: "/foo",
		err: errors.New("too much foo"),
	}

	assert.Equal(t, `could not set "/foo": too much foo`, err.Error())
}

func Test_KVDeleteError(t *testing.T) {
	err := KVDeleteError{
		key: "/foo",
		err: errors.New("too much foo"),
	}

	assert.Equal(t, `could not delete "/foo": too much foo`, err.Error())
}

func Test_NewEtcdKVS(t *testing.T) {
	ctx := context.Background()
	ksapi := &mocks.KeysAPI{}
	kvs := NewEtcdKVS(ctx, ksapi)
	assert.NotNil(t, kvs)
}

func Test_etcdKVS_Mkdir(t *testing.T) {
	ctx := context.Background()
	ksapi := &mocks.KeysAPI{}
	kvs := NewEtcdKVS(ctx, ksapi)

	opts := &client.SetOptions{
		Dir: true,
	}

	dir := "/foo"

	ksapi.On("Set", ctx, dir, "", opts).Return(nil, nil).Once()
	err := kvs.Mkdir(dir)
	assert.NoError(t, err)

	ksapi.On("Set", ctx, dir, "", opts).Return(nil, errors.New("error")).Once()
	err = kvs.Mkdir(dir)
	assert.Error(t, err)
}

func Test_etcdKVS_set(t *testing.T) {
	ctx := context.Background()
	ksapi := &mocks.KeysAPI{}
	kvs := NewEtcdKVS(ctx, ksapi)

	opts := &client.SetOptions{}

	k := "/foo"
	v := "bar"

	resp := &client.Response{
		Node: &client.Node{},
	}

	expected := &Node{Nodes: Nodes{}}

	ksapi.On("Set", ctx, k, v, opts).Return(resp, nil).Once()
	node, err := kvs.Set(k, v, nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, node)

	ksapi.On("Set", ctx, k, v, opts).Return(nil, errors.New("error")).Once()
	_, err = kvs.Set(k, v, nil)
	assert.Error(t, err)
}

func Test_etcdKVS_Rmdir(t *testing.T) {
	ctx := context.Background()
	ksapi := &mocks.KeysAPI{}
	kvs := NewEtcdKVS(ctx, ksapi)

	opts := &client.DeleteOptions{
		Dir: true,
	}

	dir := "/foo"

	ksapi.On("Delete", ctx, dir, opts).Return(nil, nil).Once()
	err := kvs.Rmdir(dir)
	assert.NoError(t, err)

	ksapi.On("Delete", ctx, dir, opts).Return(nil, errors.New("error")).Once()
	err = kvs.Rmdir(dir)
	assert.Error(t, err)
}

func Test_etcdKVS_delete(t *testing.T) {
	ctx := context.Background()
	ksapi := &mocks.KeysAPI{}
	kvs := NewEtcdKVS(ctx, ksapi)

	opts := &client.DeleteOptions{}

	key := "/foo"

	ksapi.On("Delete", ctx, key, opts).Return(nil, nil).Once()
	err := kvs.Delete(key)
	assert.NoError(t, err)

	ksapi.On("Delete", ctx, key, opts).Return(nil, errors.New("error")).Once()
	err = kvs.Delete(key)
	assert.Error(t, err)
}

func Test_etcdKVS_get(t *testing.T) {
	ctx := context.Background()
	ksapi := &mocks.KeysAPI{}
	kvs := NewEtcdKVS(ctx, ksapi)

	opts := &client.GetOptions{}

	key := "/foo"

	resp := &client.Response{
		Node: &client.Node{
			Value: "bar",
		},
	}

	recursiveResp := &client.Response{
		Node: &client.Node{
			Value: "root",
			Nodes: client.Nodes{
				&client.Node{Value: "c"},
				&client.Node{Value: "b"},
				&client.Node{Value: "a"},
			},
		},
	}

	ksapi.On("Get", ctx, key, opts).Return(resp, nil).Once()
	node, err := kvs.Get(key, nil)
	assert.NoError(t, err)
	assert.Equal(t, "bar", node.Value)

	ksapi.On("Get", ctx, key, opts).Return(nil, errors.New("error")).Once()
	_, err = kvs.Get(key, nil)
	assert.Error(t, err)

	ksapi.On("Get", ctx, key, opts).Return(resp, nil).Once()

	opts = &client.GetOptions{Recursive: true}
	ksapi.On("Get", ctx, key, opts).Return(recursiveResp, nil).Once()
	getOpts := &GetOptions{Recursive: true}
	node, err = kvs.Get(key, getOpts)
	assert.NoError(t, err)
	assert.Equal(t, "root", node.Value)
	assert.Equal(t, "c", node.Nodes[0].Value)
	assert.Equal(t, 3, len(node.Nodes))
}

func Test_newHaproxyKVS(t *testing.T) {
	mk := &mockKVS{}
	hk := NewHaproxyKVS(mk)

	assert.NotNil(t, hk)
}

func Test_haproxyKVS_init(t *testing.T) {
	mk := &mockKVS{}
	hk := NewHaproxyKVS(mk)

	mk.On("Mkdir", "/haproxy-discover/services").Return(nil).Once()
	mk.On("Mkdir", "/haproxy-discover/tcp-services").Return(nil).Once()

	err := hk.Init()
	assert.NoError(t, err)
}

func Test_haproxyKVS_domain(t *testing.T) {
	mk := &mockKVS{}
	hk := NewHaproxyKVS(mk)

	mk.On("Set", "/haproxy-discover/services/app/domain", "domain", mock.Anything).Return(&Node{}, nil).Once()
	err := hk.Domain("app", "domain")
	assert.NoError(t, err)
}

func Test_haproxyKVS_urlReg(t *testing.T) {
	mk := &mockKVS{}
	hk := NewHaproxyKVS(mk)

	mk.On("Set", "/haproxy-discover/services/app/url_reg", "regex", mock.Anything).Return(&Node{}, nil).Once()
	err := hk.URLReg("app", "regex")
	assert.NoError(t, err)
}

func Test_haproxyKVS_upstream(t *testing.T) {
	mk := &mockKVS{}
	hk := NewHaproxyKVS(mk)

	mk.On("Set", "/haproxy-discover/services/app/upstreams/nodeA", "127.0.0.1:80", mock.Anything).Return(&Node{}, nil).Once()
	err := hk.Upstream("app", "nodeA", "127.0.0.1:80")
	assert.NoError(t, err)
}

func Test_newCmKVS(t *testing.T) {
	mk := &mockKVS{}
	ck := NewCmKVS(mk, 10*time.Second)

	assert.NotNil(t, ck)
}

func Test_cmKVS_RegisterAgent(t *testing.T) {
	mk := &mockKVS{}
	ck := NewCmKVS(mk, 10*time.Second)

	node := &Node{
		ModifiedIndex: uint64(10),
	}

	mk.On("Set", "/agent/leader/agent-1", "agent-1", mock.Anything).Return(node, nil).Once()
	mi, err := ck.RegisterAgent("agent-1")

	assert.Equal(t, uint64(10), mi)
	assert.NoError(t, err)

	mk.On("Set", "/agent/leader/agent-1", "agent-1", mock.Anything).Return(nil, errors.New("fail")).Once()
	_, err = ck.RegisterAgent("agent-1")
	assert.Error(t, err)
}

func Test_cmKVS_Leader(t *testing.T) {
	mk := &mockKVS{}
	ck := NewCmKVS(mk, 10*time.Second)

	node := &Node{
		Nodes: Nodes{
			&Node{CreatedIndex: uint64(3), Value: "agent-3"},
			&Node{CreatedIndex: uint64(2), Value: "agent-2"},
			&Node{CreatedIndex: uint64(1), Value: "agent-1"},
		},
	}

	opts := &GetOptions{Recursive: true}

	mk.On("Get", "/agent/leader", opts).Return(node, nil).Once()
	leader, err := ck.Leader()

	assert.NoError(t, err)
	assert.Equal(t, 3, leader.NodeCount)
	assert.Equal(t, "agent-1", leader.Name)

	mk.On("Get", "/agent/leader", opts).Return(nil, errors.New("fail")).Once()
	_, err = ck.Leader()

	assert.Error(t, err)
}

func Test_cmKVS_Refresh(t *testing.T) {
	mk := &mockKVS{}
	ck := NewCmKVS(mk, 10*time.Second)

	node := &Node{
		ModifiedIndex: 15,
	}

	opts := &SetOptions{
		TTL:       10 * time.Second,
		PrevIndex: uint64(6),
	}

	mk.On("Set", "/agent/leader/agent-1", "agent-1", opts).Return(node, nil).Once()

	mi, err := ck.Refresh("agent-1", uint64(6))

	assert.Equal(t, uint(15), mi)
	assert.NoError(t, err)

	mk.On("Set", "/agent/leader/agent-1", "agent-1", opts).Return(nil, errors.New("fail")).Once()

	_, err = ck.Refresh("agent-1", uint64(6))
	assert.Error(t, err)
}

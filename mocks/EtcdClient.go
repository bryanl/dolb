package mocks

import "github.com/stretchr/testify/mock"

import "time"
import "github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"

type EtcdClient struct {
	mock.Mock
}

func (_m *EtcdClient) Sync(_a0 context.Context) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *EtcdClient) AutoSync(_a0 context.Context, _a1 time.Duration) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Duration) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *EtcdClient) Endpoints() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

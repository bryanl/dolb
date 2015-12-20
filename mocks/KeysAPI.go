package mocks

import "github.com/coreos/etcd/client"
import "github.com/stretchr/testify/mock"

import "github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"

type KeysAPI struct {
	mock.Mock
}

func (_m *KeysAPI) Get(ctx context.Context, key string, opts *client.GetOptions) (*client.Response, error) {
	ret := _m.Called(ctx, key, opts)

	var r0 *client.Response
	if rf, ok := ret.Get(0).(func(context.Context, string, *client.GetOptions) *client.Response); ok {
		r0 = rf(ctx, key, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *client.GetOptions) error); ok {
		r1 = rf(ctx, key, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *KeysAPI) Set(ctx context.Context, key string, value string, opts *client.SetOptions) (*client.Response, error) {
	ret := _m.Called(ctx, key, value, opts)

	var r0 *client.Response
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *client.SetOptions) *client.Response); ok {
		r0 = rf(ctx, key, value, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, *client.SetOptions) error); ok {
		r1 = rf(ctx, key, value, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *KeysAPI) Delete(ctx context.Context, key string, opts *client.DeleteOptions) (*client.Response, error) {
	ret := _m.Called(ctx, key, opts)

	var r0 *client.Response
	if rf, ok := ret.Get(0).(func(context.Context, string, *client.DeleteOptions) *client.Response); ok {
		r0 = rf(ctx, key, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *client.DeleteOptions) error); ok {
		r1 = rf(ctx, key, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *KeysAPI) Create(ctx context.Context, key string, value string) (*client.Response, error) {
	ret := _m.Called(ctx, key, value)

	var r0 *client.Response
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *client.Response); ok {
		r0 = rf(ctx, key, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, key, value)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *KeysAPI) CreateInOrder(ctx context.Context, dir string, value string, opts *client.CreateInOrderOptions) (*client.Response, error) {
	ret := _m.Called(ctx, dir, value, opts)

	var r0 *client.Response
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *client.CreateInOrderOptions) *client.Response); ok {
		r0 = rf(ctx, dir, value, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, *client.CreateInOrderOptions) error); ok {
		r1 = rf(ctx, dir, value, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *KeysAPI) Update(ctx context.Context, key string, value string) (*client.Response, error) {
	ret := _m.Called(ctx, key, value)

	var r0 *client.Response
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *client.Response); ok {
		r0 = rf(ctx, key, value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, key, value)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *KeysAPI) Watcher(key string, opts *client.WatcherOptions) client.Watcher {
	ret := _m.Called(key, opts)

	var r0 client.Watcher
	if rf, ok := ret.Get(0).(func(string, *client.WatcherOptions) client.Watcher); ok {
		r0 = rf(key, opts)
	} else {
		r0 = ret.Get(0).(client.Watcher)
	}

	return r0
}

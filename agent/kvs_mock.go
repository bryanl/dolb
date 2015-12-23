package agent

import "github.com/stretchr/testify/mock"

type mockKVS struct {
	mock.Mock
}

func (_m *mockKVS) Delete(key string) error {
	ret := _m.Called(key)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *mockKVS) Get(key string, options *GetOptions) (*Node, error) {
	ret := _m.Called(key, options)

	var r0 *Node
	if rf, ok := ret.Get(0).(func(string, *GetOptions) *Node); ok {
		r0 = rf(key, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *GetOptions) error); ok {
		r1 = rf(key, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *mockKVS) Mkdir(dir string) error {
	ret := _m.Called(dir)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(dir)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *mockKVS) Rmdir(dir string) error {
	ret := _m.Called(dir)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(dir)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func (_m *mockKVS) Set(key string, value string, options *SetOptions) (*Node, error) {
	ret := _m.Called(key, value, options)

	var r0 *Node
	if rf, ok := ret.Get(0).(func(string, string, *SetOptions) *Node); ok {
		r0 = rf(key, value, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, *SetOptions) error); ok {
		r1 = rf(key, value, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

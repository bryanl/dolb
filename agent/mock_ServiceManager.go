package agent

import "github.com/stretchr/testify/mock"

import "github.com/bryanl/dolb/kvs"
import "github.com/bryanl/dolb/service"

type MockServiceManager struct {
	mock.Mock
}

func (_m *MockServiceManager) AddUpstream(svc string, ucr UpstreamCreateRequest) error {
	ret := _m.Called(svc, ucr)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, UpstreamCreateRequest) error); ok {
		r0 = rf(svc, ucr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockServiceManager) DeleteService(svcName string) error {
	ret := _m.Called(svcName)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(svcName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockServiceManager) DeleteUpstream(svc string, upstreamID string) error {
	ret := _m.Called(svc, upstreamID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(svc, upstreamID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockServiceManager) Create(_a0 service.ServiceCreateRequest) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(service.ServiceCreateRequest) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockServiceManager) Services() ([]kvs.Service, error) {
	ret := _m.Called()

	var r0 []kvs.Service
	if rf, ok := ret.Get(0).(func() []kvs.Service); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]kvs.Service)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockServiceManager) Service(name string) (kvs.Service, error) {
	ret := _m.Called(name)

	var r0 kvs.Service
	if rf, ok := ret.Get(0).(func(string) kvs.Service); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(kvs.Service)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
